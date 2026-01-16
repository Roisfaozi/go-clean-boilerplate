package repository

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
)

type TokenRepository interface {
	StoreToken(ctx context.Context, session *model.Auth) error
	GetToken(ctx context.Context, userID, sessionID string) (*model.Auth, error)
	DeleteToken(ctx context.Context, userID, sessionID string) error
	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error
	Save(ctx context.Context, token *entity.PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*entity.PasswordResetToken, error)
	DeleteByEmail(ctx context.Context, email string) error
	DeleteExpiredResetTokens(ctx context.Context) error

	// Email Verification Token Methods
	SaveVerificationToken(ctx context.Context, token *entity.EmailVerificationToken) error
	FindVerificationToken(ctx context.Context, token string) (*entity.EmailVerificationToken, error)
	DeleteVerificationTokenByEmail(ctx context.Context, email string) error

	// Account Lockout Methods
	GetLoginAttempts(ctx context.Context, username string) (int, error)
	IncrementLoginAttempts(ctx context.Context, username string) error
	ResetLoginAttempts(ctx context.Context, username string) error
	LockAccount(ctx context.Context, username string, duration time.Duration) error
	IsAccountLocked(ctx context.Context, username string) (bool, time.Duration, error)
}
