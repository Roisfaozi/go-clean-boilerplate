package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils"
	"github.com/Roisfaozi/casbin-db/internal/utils/jwt"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service struct {
	jwtManager *jwt.JWTManager
	tokenRepo  repository.TokenRepository
	userRepo   userRepository.UserRepository
	validate   *validator.Validate
	tm         tx.WithTransactionManager
	log        *logrus.Logger
	wsManager  ws.Manager
}

func NewAuthUsecase(
	jwtManager *jwt.JWTManager,
	tokenRepo repository.TokenRepository,
	userRepo userRepository.UserRepository,
	validate *validator.Validate,
	tm tx.WithTransactionManager,
	log *logrus.Logger,
	wsManager ws.Manager,
) AuthUseCase {
	return &Service{
		jwtManager: jwtManager,
		tokenRepo:  tokenRepo,
		userRepo:   userRepo,
		validate:   validate,
		tm:         tm,
		log:        log,
		wsManager:  wsManager,
	}
}

func (s *Service) generateAndStoreTokenPair(user *entity.User) (string, string, error) {
	sessionID := uuid.New().String()

	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate token pair: %w", err)
	}

	session := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtManager.GetRefreshTokenDuration()),
	}

	if err := s.tokenRepo.StoreToken(context.Background(), session); err != nil {
		s.log.WithError(err).Error("Failed to store session in Redis")
		return "", "", fmt.Errorf("failed to store session: %w", err)
	}

	return accessToken, refreshToken, nil
}

// Login handles user login and returns access and refresh tokens
func (s *Service) Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error) {
	if err := s.validate.Struct(request); err != nil {
		return nil, "", err
	}

	var user *entity.User
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = s.userRepo.FindByUsername(txCtx, request.Username)
		if err != nil {
			return errors.New("invalid credentials")
		}

		if !utils.CheckPasswordHash(request.Password, user.Password) {
			return errors.New("invalid credentials")
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	accessToken, refreshToken, err := s.generateAndStoreTokenPair(user)
	if err != nil {
		return nil, "", err
	}

	// Broadcast login event via WebSocket
	notification := map[string]string{
		"type":    "user_login",
		"user_id": user.ID,
		"message": fmt.Sprintf("User '%s' has just logged in.", user.Name),
		"time":    time.Now().Format(time.RFC3339),
	}
	notificationJSON, _ := json.Marshal(notification)
	s.wsManager.BroadcastToChannel("global_notifications", notificationJSON)

	loginResponse := &model.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	}

	return loginResponse, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, "", err
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, "", err
	}

	// Revoke the old session
	if err := s.RevokeToken(ctx, claims.UserID, claims.SessionID); err != nil {
		s.log.WithError(err).Warn("Failed to revoke old session during refresh")
	}

	newAccessToken, newRefreshToken, err := s.generateAndStoreTokenPair(user)
	if err != nil {
		return nil, "", err
	}

	tokenResponse := &model.TokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
	}

	return tokenResponse, newRefreshToken, nil
}

// ValidateAccessToken validates an access token and returns its claims
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.validateSession(claims, tokenString)
}

// ValidateRefreshToken validates a refresh token and returns its claims
func (s *Service) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return s.validateSession(claims, tokenString)
}

// validateSession checks if the session associated with the token is valid.
func (s *Service) validateSession(claims *jwt.Claims, tokenString string) (*Claims, error) {
	// Check if token exists in Redis
	savedSession, err := s.tokenRepo.GetToken(context.Background(), claims.UserID, claims.SessionID)
	if err != nil {
		// Considering DB or Redis errors as a reason for revocation check failure
		return nil, ErrTokenRevoked
	}

	// If no session is found in Redis, the token is considered revoked or invalid
	if savedSession == nil {
		return nil, ErrTokenRevoked
	}

	// Ensure the token being validated matches the one in the session
	isAccessToken := savedSession.AccessToken == tokenString
	isRefreshToken := savedSession.RefreshToken == tokenString
	if !isAccessToken && !isRefreshToken {
		return nil, ErrTokenRevoked
	}

	customClaims := &Claims{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
	}

	return customClaims, nil
}

// Verify verifies the user's session
func (s *Service) Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error) {
	return s.tokenRepo.GetToken(ctx, userID, sessionID)
}

func (s *Service) RevokeToken(ctx context.Context, userID, sessionID string) error {
	s.log.Infof("Revoking token for user %s with session %s", userID, sessionID)
	return s.tokenRepo.DeleteToken(ctx, userID, sessionID)
}

func (s *Service) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	s.log.Infof("Getting all sessions for user %s", userID)
	return s.tokenRepo.GetUserSessions(ctx, userID)

}

func (s *Service) RevokeAllSessions(ctx context.Context, userID string) error {
	s.log.Infof("Revoking all sessions for user %s", userID)
	return s.tokenRepo.RevokeAllSessions(ctx, userID)
}

// GenerateAccessToken generates a new access token for the user
func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	accessToken, _, err := s.jwtManager.GenerateTokenPair(user.ID, uuid.NewString())
	return accessToken, err
}

// GenerateRefreshToken generates a new refresh token for the user
func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	_, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, uuid.NewString())
	return refreshToken, err
}
