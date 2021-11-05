package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
	"github.com/caarlos0/env"
	"github.com/rs/cors"

	"github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"
)

type environmentVariables struct {
	StravaClientID     string `env: "STRAVA_CLIENT_ID, required"`
	StravaClientSecret string `env: "STRAVA_CLIENT_SECRET, required"`
	StravaRefreshToken string `env: "STRAVA_REFRESH_TOKEN, required"`
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

func main() {

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	log := logrus.NewEntry(logger)
	log.WithField("component", "easy-strava-upload")

	log.Info("Starting Easy-Strava-Upload Application")

	log.Fatal(http.ListenAndServe(":8090", cors.Default().Handler(api.NewServer())))

}
