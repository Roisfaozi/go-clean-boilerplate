package config

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type silentLogger struct{}

func (l silentLogger) Printf(ctx context.Context, format string, v ...interface{}) {}

func NewRedisConfig(cfg *AppConfig, log *logrus.Logger) *redis.Client {
	// Silence redis internal logger in test mode to avoid handshake warnings
	if gin.Mode() == gin.TestMode {
		redis.SetLogger(silentLogger{})
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
		// Force RESP2 for compatibility across various environments and versions
		Protocol:        2,
		DisableIdentity: true,
	})

	log.Infof("Redis connection established: %s", cfg.Redis.Addr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	return redisClient
}
