package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
	"github.com/adiazny/easy-strava-upload/internal/pkg/store"
	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
	"github.com/caarlos0/env"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	oStrava "github.com/markbates/goth/providers/strava"
	"github.com/rs/cors"

	"github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"
)

const (
	appName = "easy-strava-upload"
)

type environmentVariables struct {
	StravaClientID     string `env:"STRAVA_CLIENT_ID,required"`
	StravaClientSecret string `env:"STRAVA_CLIENT_SECRET,required"`
	StravaCallback     string `env:"STRAVA_CALLBACK,required"`
	StravaRefreshToken string `env:"STRAVA_REFRESH_TOKEN"`
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

func newServer(log *logrus.Entry, provider strava.Provider, provIndex *api.ProviderIndex) *api.Server {
	s := &api.Server{
		Log:            log,
		Router:         http.NewServeMux(),
		StravaProvider: &provider,
		ProviderIndex:  provIndex,
	}

	s.Routes()

	log.Info("Easy-Strava-Upload Application Started")
	return s
}

func main() {

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	log := logrus.NewEntry(logger)
	log.WithField("component", appName).Info("starting up")
	defer log.Info("Shutting down")

	envVars, err := setup()
	if err != nil {
		log.WithError(err).Error()
	}

	key := "Secret-session-key" // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30        // 30 days
	isProd := false             // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store

	// Registers a list of available providers for use with Goth. The Strava provider is added only at the moment.
	goth.UseProviders(
		oStrava.New(envVars.StravaClientID, envVars.StravaClientSecret, envVars.StravaCallback, "profile:write"),
	)

	provMap := make(map[string]string)
	provMap["strava"] = "Strava"

	var keys []string
	for key := range provMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	providerIndex := &api.ProviderIndex{Providers: keys, ProvidersMap: provMap}

	stravaProvider := makeStravaProvider(log, envVars)

	/*
		athleteAccess := &strava.AthleteAccessInfo{
			ID:           strava.AthleteID,
			Username:     strava.AthleteUsername,
			RefreshToken: envVars.StravaRefreshToken,
			AccessToken:  "",
			ExpiresAt:    0,
			ExpiresIn:    0,
		}
		log.Infof("AthleteAccessInfo ", athleteAccess)

		athleteAccessInfoJSON, err := json.Marshal(athleteAccess)
		if err != nil {
			log.Infof("Failed to load data into redis: %w")
			return
		}

		stravaProvider.Redis.Store(strava.AthleteID, athleteAccessInfoJSON)
	*/

	log.Fatal(http.ListenAndServe(":8090", cors.Default().Handler(newServer(log, *stravaProvider, providerIndex))))

}
