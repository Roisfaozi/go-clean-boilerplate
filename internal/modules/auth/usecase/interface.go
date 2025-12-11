package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

type AuthUseCase interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(token string) (*jwt.Claims, error)  // Updated return type
	ValidateRefreshToken(token string) (*jwt.Claims, error) // Updated return type
	RevokeToken(ctx context.Context, userID, sessionID string) error

	Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error)
	Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error)

	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error
}
