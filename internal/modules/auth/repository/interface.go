package repository

import (
	"context"
	"time"
)

type TokenRepository interface {
	StoreToken(ctx context.Context, userID string, token string, expiration time.Duration) error
	GetToken(ctx context.Context, userID string) (string, error)
	DeleteToken(ctx context.Context, userID string) error
}
