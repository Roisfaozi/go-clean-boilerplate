package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
)

var (
	ErrInvalidCredentials = exception.ErrUnauthorized
	ErrInvalidToken       = exception.ErrUnauthorized
	ErrExpiredToken       = exception.ErrUnauthorized
	ErrTokenRevoked       = exception.ErrUnauthorized
	ErrInvalidResetToken  = exception.ErrBadRequest
)

type AuthUseCase interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(token string) (*jwt.Claims, error)
	ValidateRefreshToken(token string) (*jwt.Claims, error)
	RevokeToken(ctx context.Context, userID, sessionID string) error

	Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error)
	Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error)

	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error

	// Password Recovery
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}
