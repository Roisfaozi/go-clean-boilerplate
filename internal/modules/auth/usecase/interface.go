package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
)

var (
	ErrInvalidCredentials       = exception.ErrUnauthorized
	ErrInvalidToken             = exception.ErrUnauthorized
	ErrExpiredToken             = exception.ErrUnauthorized
	ErrTokenRevoked             = exception.ErrUnauthorized
	ErrInvalidResetToken        = exception.ErrBadRequest
	ErrInvalidVerificationToken = exception.ErrBadRequest
	ErrAlreadyVerified          = exception.ErrBadRequest
	ErrAccountSuspended         = exception.ErrForbidden
	ErrAccountLocked            = exception.ErrForbidden
)

type AuthUseCase interface {
	GenerateAccessToken(user *entity.User) (string, error)
	GenerateRefreshToken(user *entity.User) (string, error)
	ValidateAccessToken(token string) (*jwt.Claims, error)
	ValidateRefreshToken(token string) (*jwt.Claims, error)
	RevokeToken(ctx context.Context, userID, sessionID string) error

	Register(ctx context.Context, request model.RegisterRequest) (*model.LoginResponse, string, error)
	Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error)
	Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error)

	GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error)
	RevokeAllSessions(ctx context.Context, userID string) error

	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Email Verification
	RequestVerification(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, token string) error
}
