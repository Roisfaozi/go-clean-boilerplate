package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

// Claims represents the JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

type AuthUseCase interface {
	// Token Management
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(token string) (*Claims, error)
	ValidateRefreshToken(token string) (*Claims, error)
	RevokeToken(ctx context.Context, userID, sessionID string) error

	// Authentication
	Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error)
	Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error)

	// Session Management
	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error
}
