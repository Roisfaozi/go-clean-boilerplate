package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
)

type TokenRepository interface {
	StoreToken(ctx context.Context, session *model.Auth) error
	GetToken(ctx context.Context, userID, sessionID string) (*model.Auth, error)
	DeleteToken(ctx context.Context, userID, sessionID string) error
	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error
}
