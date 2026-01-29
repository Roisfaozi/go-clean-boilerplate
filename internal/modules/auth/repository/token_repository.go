package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type tokenRepositoryRedis struct {
	client *redis.Client
	log    *logrus.Logger
	db     *gorm.DB
}

func (r *tokenRepositoryRedis) Save(ctx context.Context, token *entity.PasswordResetToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *tokenRepositoryRedis) FindByToken(ctx context.Context, token string) (*entity.PasswordResetToken, error) {
	var resetToken entity.PasswordResetToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

func (r *tokenRepositoryRedis) DeleteByEmail(ctx context.Context, email string) error {
	if err := r.db.WithContext(ctx).Where("email = ?", email).Delete(&entity.PasswordResetToken{}).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to delete reset token by email")
		return err
	}
	return nil
}

func (r *tokenRepositoryRedis) DeleteExpiredResetTokens(ctx context.Context) error {
	// Deletes tokens where expires_at < NOW()
	if err := r.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&entity.PasswordResetToken{}).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to delete expired reset tokens")
		return err
	}
	return nil
}

func NewTokenRepositoryRedis(client *redis.Client, log *logrus.Logger, db *gorm.DB) TokenRepository {
	return &tokenRepositoryRedis{
		client: client,
		log:    log,
		db:     db,
	}
}

func (r *tokenRepositoryRedis) StoreToken(ctx context.Context, session *model.Auth) error {
	key := r.getSessionKey(session.UserID, session.ID)

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

func (r *tokenRepositoryRedis) DeleteToken(ctx context.Context, userID, sessionID string) error {
	key := r.getSessionKey(userID, sessionID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.log.WithError(err).Error("Failed to delete session from Redis")
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

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

func (r *tokenRepositoryRedis) getSessionKey(userID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", userID, sessionID)
}

// Email Verification Token Methods

func (r *tokenRepositoryRedis) SaveVerificationToken(ctx context.Context, token *entity.EmailVerificationToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *tokenRepositoryRedis) FindVerificationToken(ctx context.Context, token string) (*entity.EmailVerificationToken, error) {
	var verificationToken entity.EmailVerificationToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&verificationToken).Error
	if err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (r *tokenRepositoryRedis) DeleteVerificationTokenByEmail(ctx context.Context, email string) error {
	if err := r.db.WithContext(ctx).Where("email = ?", email).Delete(&entity.EmailVerificationToken{}).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to delete verification token by email")
		return err
	}
	return nil
}

// Account Lockout Methods

func (r *tokenRepositoryRedis) getAttemptsKey(username string) string {
	return fmt.Sprintf("auth:attempts:%s", username)
}

func (r *tokenRepositoryRedis) getLockedKey(username string) string {
	return fmt.Sprintf("auth:locked:%s", username)
}

func (r *tokenRepositoryRedis) GetLoginAttempts(ctx context.Context, username string) (int, error) {
	key := r.getAttemptsKey(username)
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to get login attempts")
		return 0, err
	}
	return val, nil
}

func (r *tokenRepositoryRedis) IncrementLoginAttempts(ctx context.Context, username string) (int, error) {
	key := r.getAttemptsKey(username)
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 1*time.Hour) // Reset attempts counter after 1 hour of inactivity
	_, err := pipe.Exec(ctx)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to increment login attempts")
		return 0, err
	}
	return int(incr.Val()), nil
}

func (r *tokenRepositoryRedis) ResetLoginAttempts(ctx context.Context, username string) error {
	key := r.getAttemptsKey(username)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to reset login attempts")
		return err
	}
	return nil
}

func (r *tokenRepositoryRedis) LockAccount(ctx context.Context, username string, duration time.Duration) error {
	key := r.getLockedKey(username)
	if err := r.client.Set(ctx, key, "locked", duration).Err(); err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to lock account")
		return err
	}
	return nil
}

func (r *tokenRepositoryRedis) IsAccountLocked(ctx context.Context, username string) (bool, time.Duration, error) {
	key := r.getLockedKey(username)
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to check account lock status")
		return false, 0, err
	}

	if ttl <= 0 { // Key does not exist (TTL -2) or no expire (TTL -1) but logic says locked keys always expire
		return false, 0, nil
	}

	return true, ttl, nil
}
