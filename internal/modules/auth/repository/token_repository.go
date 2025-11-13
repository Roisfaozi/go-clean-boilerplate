package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type tokenRepositoryRedis struct {
	client *redis.Client
	log    *logrus.Logger
}

// NewTokenRepositoryRedis creates a new Redis-based token repository
func NewTokenRepositoryRedis(client *redis.Client, log *logrus.Logger) TokenRepository {
	return &tokenRepositoryRedis{
		client: client,
		log:    log,
	}
}

// StoreToken stores a token with session information in Redis
func (r *tokenRepositoryRedis) StoreToken(ctx context.Context, userID string, token string, expiration time.Duration) error {
	sessionID := uuid.NewString()

	// Store the token in Redis with the session ID as part of the key
	key := r.getSessionKey(userID, sessionID)
	err := r.client.Set(ctx, key, token, expiration).Err()
	if err != nil {
		r.log.WithError(err).Error("Failed to store token in Redis")
		return fmt.Errorf("failed to store token: %w", err)
	}

	return nil
}

// GetToken retrieves a token by user ID and session ID
func (r *tokenRepositoryRedis) GetToken(ctx context.Context, userID, sessionID string) (*model.Auth, error) {
	key := r.getSessionKey(userID, sessionID)
	token, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		r.log.WithError(err).Error("Failed to get token from Redis")
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Create a minimal session object with the token
	session := &model.Auth{
		ID:           sessionID,
		UserID:       userID,
		SessionID:    sessionID,
		AccessToken:  token,                          // Assuming it's an access token by default
		RefreshToken: "",                             // This would be set when storing a refresh token
		ExpiresAt:    time.Now().Add(time.Hour * 24), // Default expiration
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return session, nil
}

// DeleteToken removes a session by user ID and session ID
func (r *tokenRepositoryRedis) DeleteToken(ctx context.Context, userID, sessionID string) error {
	key := r.getSessionKey(userID, sessionID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.log.WithError(err).Error("Failed to delete session from Redis")
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *tokenRepositoryRedis) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	pattern := r.getSessionKey(userID, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.log.WithError(err).Error("Failed to get user sessions")
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessions []*model.Auth
	for _, key := range keys {
		// Extract session ID from the key
		sessionID := key[len(r.getSessionKey(userID, "")):]

		token, err := r.client.Get(ctx, key).Result()
		if err != nil {
			r.log.WithError(err).WithField("key", key).Warn("Failed to get token data")
			continue
		}

		session := &model.Auth{
			ID:           sessionID,
			UserID:       userID,
			SessionID:    sessionID,
			AccessToken:  token,                          // Assuming it's an access token
			RefreshToken: "",                             // This would be set when storing a refresh token
			ExpiresAt:    time.Now().Add(time.Hour * 24), // Default expiration
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// RevokeAllSessions revokes all active sessions for a user
func (r *tokenRepositoryRedis) RevokeAllSessions(ctx context.Context, userID string) error {
	pattern := r.getSessionKey(userID, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.log.WithError(err).Error("Failed to get user sessions for revocation")
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			r.log.WithError(err).Error("Failed to revoke user sessions")
			return fmt.Errorf("failed to revoke user sessions: %w", err)
		}
	}

	return nil
}

// getSessionKey generates a Redis key for session storage
func (r *tokenRepositoryRedis) getSessionKey(userID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", userID, sessionID)
}
