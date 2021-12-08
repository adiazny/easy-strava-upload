package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/adiazny/easy-strava-upload/app/services/strava-api/handlers"
	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
	"github.com/adiazny/easy-strava-upload/internal/pkg/store"
	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
	"github.com/ardanlabs/conf/v2"
	"github.com/caarlos0/env"

	"github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	appName = "easy-strava-upload"
)

var build = "develop"

func main() {

	// Construct the application logger.
	log, err := initLogger("STRAVA-API")
	if err != nil {
		fmt.Println("error contructing the application logger", err)
		os.Exit(1)
	}

	defer log.Sync()

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}

	// envVars, err := setup()
	// if err != nil {
	// 	log.WithError(err).Error()
	// }

	// stravaProvider := makeStravaProvider(log, envVars)

	// athleteAccess := &strava.AthleteAccessInfo{
	// 	ID:           strava.AthleteID,
	// 	Username:     strava.AthleteUsername,
	// 	RefreshToken: envVars.StravaRefreshToken,
	// 	AccessToken:  "",
	// 	ExpiresAt:    0,
	// 	ExpiresIn:    0,
	// }
	// log.Infof("AthleteAccessInfo ", athleteAccess)

	// athleteAccessInfoJSON, err := json.Marshal(athleteAccess)
	// if err != nil {
	// 	log.Infof("Failed to load data into redis: %w")
	// 	return
	// }

	// stravaProvider.Redis.Store(strava.AthleteID, athleteAccessInfoJSON)

	// log.Fatal(http.ListenAndServe(":8090", cors.Default().Handler(newServer(log, *stravaProvider))))

}

func run(log *zap.SugaredLogger) error {
	// =========================================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// based on what is available either by the machine or kubernetes resource quotas.
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// Define a configuration struct literal using ArdanLab's conf library
	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	// Parse external configuration and environement variables
	const prefix = "STRAVA" //optional
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// =========================================================================
	// App Starting

	log.Infow("starting service", "version", build)

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	expvar.NewString("build").Set(build)

	// =========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This includes the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugMux(build, log)

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the mux for the API calls.
	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Log:      log,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}

		return nil
	}
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}

type environmentVariables struct {
	StravaClientID     string `env:"STRAVA_CLIENT_ID,required"`
	StravaClientSecret string `env:"STRAVA_CLIENT_SECRET,required"`
	StravaRefreshToken string `env:"STRAVA_REFRESH_TOKEN,required"`
	RedisAddress       string `env:"REDIS_ADDRESS,required"`
	RedisPassword      string `env:"REDIS_PASSWORD,required"`
	RedisDB            int    `env:"REDIS_DB,required"`
}

func setup() (envVars *environmentVariables, err error) {
	_, err = maxprocs.Set()
	if err != nil {
		return nil, fmt.Errorf("Error setting GOMAXPROCS %w", err)
	}

	envVars = &environmentVariables{}

	err = env.Parse(envVars)
	if err != nil {
		return nil, fmt.Errorf("Error parsing environmenet varilables %w", err)
	}

	return envVars, nil

}

func makeStravaProvider(log *logrus.Entry, envVars *environmentVariables) *strava.Provider {
	stravaConfig := &strava.Config{
		StravaClientID:     envVars.StravaClientID,
		StravaClientSecret: envVars.StravaClientSecret,
		StravaRefreshToken: envVars.StravaRefreshToken,
	}

	redisConfig := &store.Config{
		Addr:     envVars.RedisAddress,
		Password: envVars.RedisPassword,
		DB:       envVars.RedisDB,
	}

	redisDB := store.NewClient(log, redisConfig)

	return strava.NewProvider(log, "strava", stravaConfig, redisDB)
}

func newServer(log *logrus.Entry, provider strava.Provider) *api.Server {
	s := &api.Server{
		Log:            log,
		Router:         http.NewServeMux(),
		StravaProvider: &provider,
	}

	s.Routes()

	log.Info("Easy-Strava-Upload Application Started")
	return s
}
