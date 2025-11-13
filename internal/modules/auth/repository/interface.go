package repository

import (
	"context"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
)

// TokenRepository defines the interface for token and session management
type TokenRepository interface {
	// Basic token operations
	StoreToken(ctx context.Context, userID string, token string, expiration time.Duration) error
	GetToken(ctx context.Context, userID, sessionID string) (*model.Auth, error)
	DeleteToken(ctx context.Context, userID, sessionID string) error

	// Enhanced session management
	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error
}
