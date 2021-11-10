package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/adiazny/easy-strava-upload/internal/pkg/api"
	"github.com/adiazny/easy-strava-upload/internal/pkg/store"
	"github.com/adiazny/easy-strava-upload/internal/pkg/strava"
	"github.com/caarlos0/env"
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

	stravaProvider := makeStravaProvider(log, envVars)

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

	log.Fatal(http.ListenAndServe(":8090", cors.Default().Handler(newServer(log, *stravaProvider))))

}
