package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
	"github.com/caarlos0/env"
	"github.com/rs/cors"

	"github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"
)

type environmentVariables struct {
	StravaClientID     string `env:"STRAVA_CLIENT_ID,required"`
	StravaClientSecret string `env:"STRAVA_CLIENT_SECRET,required"`
	StravaRefreshToken string `env:"STRAVA_REFRESH_TOKEN,required"`
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
	log.Infof("EnvVars: %s", envVars.StravaClientID)

	config := &strava.Config{
		StravaClientID:     envVars.StravaClientID,
		StravaClientSecret: envVars.StravaClientSecret,
		StravaRefreshToken: envVars.StravaRefreshToken,
	}
	log.Infof("Config: %s", config.StravaClientID)
	return strava.NewProvider(log, "strava", config)
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

func main() {

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	log := logrus.NewEntry(logger)
	log.WithField("component", "easy-strava-upload").Info("starting up")
	defer log.Info("Shutting down")

	envVars, err := setup()
	if err != nil {
		log.WithError(err).Error()
	}

	stravaProvider := makeStravaProvider(log, envVars)

	log.Fatal(http.ListenAndServe(":8090", cors.Default().Handler(newServer(log, *stravaProvider))))

}
