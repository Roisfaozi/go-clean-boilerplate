package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type tokenRepositoryRedis struct {
	client *redis.Client
	log    *logrus.Logger
}

func NewTokenRepositoryRedis(client *redis.Client, log *logrus.Logger) TokenRepository {
	return &tokenRepositoryRedis{
		client: client,
		log:    log,
	}
}

func (r *tokenRepositoryRedis) StoreToken(ctx context.Context, userID string, token string, expiration time.Duration) error {
	key := r.getKey(userID)
	err := r.client.Set(ctx, key, token, expiration).Err()
	if err != nil {
		r.log.WithError(err).Error("Failed to store token in Redis")
		return fmt.Errorf("failed to store token: %w", err)
	}
	return nil
}

func (r *tokenRepositoryRedis) GetToken(ctx context.Context, userID string) (string, error) {
	key := r.getKey(userID)
	token, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		r.log.WithError(err).Error("Failed to get token from Redis")
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token, nil
}

func (r *tokenRepositoryRedis) DeleteToken(ctx context.Context, userID string) error {
	key := r.getKey(userID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.log.WithError(err).Error("Failed to delete token from Redis")
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

func (r *tokenRepositoryRedis) getKey(userID string) string {
	return fmt.Sprintf("user:%s:token", userID)
}
