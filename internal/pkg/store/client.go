package store

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

type Redis struct {
	Log    *logrus.Entry
	Client *redis.Client
}

func (redis *Redis) Store(athleteID string, bytes []byte) error {
	err := redis.Client.Set(context.Background(), string(athleteID), bytes, 0).Err()
	if err != nil {
		return fmt.Errorf("Error storing value %v for key %s", bytes, athleteID)
	}

	return nil
}

func (redis *Redis) loadData() {
}

func NewClient(log *logrus.Entry, config *Config) *Redis {
	return &Redis{
		Log: log,
		Client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
		}),
	}
}
