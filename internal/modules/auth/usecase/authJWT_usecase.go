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
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Service struct {
	config        Config
	tokenRepo     repository.TokenRepository
	tokenDuration map[TokenType]time.Duration
	userRepo      userRepository.UserRepository
	validate      *validator.Validate
	tm            tx.WithTransactionManager
	log           *logrus.Logger
	wsManager     ws.Manager
}

type Config interface {
	GetAccessTokenSecret() string
	GetRefreshTokenSecret() string
	GetAccessTokenDuration() time.Duration
	GetRefreshTokenDuration() time.Duration
}

func NewService(
	config Config,
	tokenRepo repository.TokenRepository,
	userRepo userRepository.UserRepository,
	validate *validator.Validate,
	tm tx.WithTransactionManager,
	log *logrus.Logger,
	wsManager ws.Manager,
) AuthUseCase {
	return &Service{
		config:    config,
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		validate:  validate,
		tm:        tm,
		log:       log,
		wsManager: wsManager,
		tokenDuration: map[TokenType]time.Duration{
			AccessToken:  config.GetAccessTokenDuration(),
			RefreshToken: config.GetRefreshTokenDuration(),
		},
	}
}

func (s *Service) generateTokenPair(user *entity.User) (string, string, error) {
	sessionID := uuid.New().String()

	accessToken, err := s.generateToken(user, AccessToken, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(user, RefreshToken, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.storeSession(user.ID, sessionID, accessToken, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) storeSession(userID, sessionID, accessToken, refreshToken string) error {
	session := &model.Auth{
		ID:           sessionID,
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.GetRefreshTokenDuration()),
	}

	err := s.tokenRepo.StoreToken(context.Background(), session)
	if err != nil {
		s.log.WithError(err).Error("Failed to store session in Redis")
		return fmt.Errorf("failed to store session: %w", err)
	}
	return nil
}

// generateToken generates a JWT token for the given user and token type
func (s *Service) generateToken(user *entity.User, tokenType TokenType, sessionID string) (string, error) {
	var secret string
	var expiresIn time.Duration

	switch tokenType {
	case AccessToken:
		secret = s.config.GetAccessTokenSecret()
		expiresIn = s.tokenDuration[AccessToken]
	case RefreshToken:
		secret = s.config.GetRefreshTokenSecret()
		expiresIn = s.tokenDuration[RefreshToken]
	default:
		return "", fmt.Errorf("unsupported token type: %s", tokenType)
	}

	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        sessionID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		s.log.WithError(err).Error("Failed to sign token")
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// GenerateAccessToken generates a new access token for the user
func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	return s.generateToken(user, AccessToken, uuid.NewString())
}

// GenerateRefreshToken generates a new refresh token for the user
func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	return s.generateToken(user, RefreshToken, uuid.NewString())
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

	accessToken, refreshToken, err := s.generateTokenPair(user)
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

	newAccessToken, newRefreshToken, err := s.generateTokenPair(user)
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
	return s.validateToken(tokenString, s.config.GetAccessTokenSecret())
}

// ValidateRefreshToken validates a refresh token and returns its claims
func (s *Service) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.GetRefreshTokenSecret())
}

// Verify verifies the user's session
func (s *Service) Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error) {
	return s.tokenRepo.GetToken(ctx, userID, sessionID)
}

func (s *Service) validateToken(tokenString string, secret string) (*Claims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token exists in Redis
	savedSession, err := s.tokenRepo.GetToken(context.Background(), claims.Subject, claims.ID)
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
		UserID:    claims.Subject,
		SessionID: claims.ID,
	}

	return customClaims, nil
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

// GenerateTestToken is a helper function for testing purposes.
// It should not be used in production code.
func GenerateTestToken(userID, sessionID, secret string, expiry time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		ID:        sessionID,
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
