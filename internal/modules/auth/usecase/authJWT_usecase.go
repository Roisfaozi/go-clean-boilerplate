package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
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
	tm            tx.TransactionManager
	log           *logrus.Logger
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
	tm tx.TransactionManager,
	log *logrus.Logger,
) *Service {
	return &Service{
		config:    config,
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		validate:  validate,
		tm:        tm,
		log:       log,
		tokenDuration: map[TokenType]time.Duration{
			AccessToken:  config.GetAccessTokenDuration(),
			RefreshToken: config.GetRefreshTokenDuration(),
		},
	}
}

// generateToken generates a JWT token for the given user and token type
func (s *Service) generateToken(user *entity.User, tokenType TokenType) (string, error) {
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

	sessionID := uuid.New().String()
	now := time.Now()

	// Create claims with expiration
	expiresAt := now.Add(expiresIn)
	claims := &jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
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

	// Store session information
	session := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		SessionID:    sessionID,
		AccessToken:  "",
		RefreshToken: "",
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if tokenType == AccessToken {
		session.AccessToken = tokenString
	} else {
		session.RefreshToken = tokenString
	}

	// Store session in Redis
	err = s.tokenRepo.StoreToken(
		context.Background(),
		user.ID,
		tokenString,
		expiresIn,
	)
	if err != nil {
		s.log.WithError(err).Error("Failed to store token in Redis")
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return tokenString, nil
}

// GenerateAccessToken generates a new access token for the user
func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	return s.generateToken(user, AccessToken)
}

// GenerateRefreshToken generates a new refresh token for the user
func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	return s.generateToken(user, RefreshToken)
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

		sessionID := uuid.NewString()
		user.Token = sessionID
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := s.GenerateRefreshToken(user)
	if err != nil {
		return nil, "", err
	}

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

	user, err := s.userRepo.FindByID(ctx, claims.Subject)
	if err != nil {
		return nil, "", err
	}

	newAccessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	newRefreshToken, err := s.GenerateRefreshToken(user)
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
func (s *Service) ValidateAccessToken(tokenString string) (*jwt.RegisteredClaims, error) {
	return s.validateToken(tokenString, s.config.GetAccessTokenSecret())
}

// ValidateRefreshToken validates a refresh token and returns its claims
func (s *Service) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	return s.validateToken(tokenString, s.config.GetRefreshTokenSecret())
}

// Verify verifies the user's session
func (s *Service) Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error) {
	var auth *model.Auth
	err := s.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.userRepo.FindByID(txCtx, userID)
		if err != nil {

			return exception.ErrNotFound
		}

		if user.Token == "" || user.Token != sessionID {
			s.log.Warnf("User token mismatch or empty for user : %s", userID)
			return exception.ErrUnauthorized
		}

		auth = &model.Auth{ID: user.ID}
		return nil
	})

	return auth, err
}

func (s *Service) validateToken(tokenString string, secret string) (*jwt.RegisteredClaims, error) {
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
	savedToken, err := s.tokenRepo.GetToken(context.Background(), claims.Subject, claims.ID)
	if err != nil {
		return nil, ErrTokenRevoked
	}

	// If no token found in Redis or the tokens don't match
	if savedToken == nil || (savedToken.AccessToken != tokenString && savedToken.RefreshToken != tokenString) {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}
