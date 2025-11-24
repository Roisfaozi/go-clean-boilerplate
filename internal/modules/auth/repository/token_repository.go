package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
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

// StoreToken stores a session in Redis
func (r *tokenRepositoryRedis) StoreToken(ctx context.Context, session *model.Auth) error {
	key := r.getSessionKey(session.UserID, session.ID)

	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		r.log.WithError(err).Error("Failed to marshal session to JSON")
		return fmt.Errorf("failed to store session: %w", err)
	}

	expiration := time.Until(session.ExpiresAt)
	err = r.client.Set(ctx, key, sessionJSON, expiration).Err()
	if err != nil {
		r.log.WithError(err).Error("Failed to store session in Redis")
		return fmt.Errorf("failed to store session: %w", err)
	}

	return nil
}

// GetToken retrieves a session by user ID and session ID
func (r *tokenRepositoryRedis) GetToken(ctx context.Context, userID, sessionID string) (*model.Auth, error) {
	key := r.getSessionKey(userID, sessionID)
	sessionJSON, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		r.log.WithError(err).Error("Failed to get session from Redis")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session model.Auth
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		r.log.WithError(err).Error("Failed to unmarshal session from JSON")
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
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
		r.log.WithError(err).Error("Failed to get user session keys")
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessions []*model.Auth
	for _, key := range keys {
		sessionJSON, err := r.client.Get(ctx, key).Result()
		if err != nil {
			r.log.WithError(err).WithField("key", key).Warn("Failed to get session data for key")
			continue
		}

		var session model.Auth
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			r.log.WithError(err).WithField("key", key).Warn("Failed to unmarshal session data")
			continue
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// RevokeAllSessions revokes all active sessions for a user
func (r *tokenRepositoryRedis) RevokeAllSessions(ctx context.Context, userID string) error {
	pattern := r.getSessionKey(userID, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.log.WithError(err).Error("Failed to get user sessions for revocation")
		return fmt.Errorf("failed to get user sessions for revocation: %w", err)
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
