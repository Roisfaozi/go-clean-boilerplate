package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func NewRedisConfig(cfg *AppConfig, log *logrus.Logger) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s", cfg.Redis.Addr),
		Password: cfg.Redis.Password,
		DB:       0,
	})

	log.Infof("Redis connection established: %s", cfg.Redis.Addr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return redisClient
}
