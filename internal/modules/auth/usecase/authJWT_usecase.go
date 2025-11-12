package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
	ErrTokenRevoked = errors.New("token has been revoked") // Add this line

)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

type Service struct {
	config        Config
	tokenRepo     repository.TokenRepository
	tokenDuration map[TokenType]time.Duration
}

type Config interface {
	GetAccessTokenSecret() string
	GetRefreshTokenSecret() string
	GetAccessTokenDuration() time.Duration
	GetRefreshTokenDuration() time.Duration
}

func NewService(config Config, tokenRepo repository.TokenRepository) *Service {
	return &Service{
		config:    config,
		tokenRepo: tokenRepo,
		tokenDuration: map[TokenType]time.Duration{
			AccessToken:  config.GetAccessTokenDuration(),
			RefreshToken: config.GetRefreshTokenDuration(),
		},
	}
}

func (s *Service) GenerateAccessToken(user *entity.User) (string, error) {
	token, err := s.generateToken(user, AccessToken)
	if err != nil {
		return "", err
	}

	err = s.tokenRepo.StoreToken(
		context.Background(),
		user.ID,
		token,
		s.tokenDuration[AccessToken],
	)
	if err != nil {
		return "", err
	}

	return token, err
}

func (s *Service) GenerateRefreshToken(user *entity.User) (string, error) {
	token, err := s.generateToken(user, RefreshToken)
	if err != nil {
		return "", err
	}

	// Store the refresh token in Redis with longer expiration
	err = s.tokenRepo.StoreToken(
		context.Background(),
		user.ID+":refresh",
		token,
		s.tokenDuration[RefreshToken],
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) generateToken(user *entity.User, tokenType TokenType) (string, error) {
	now := time.Now()

	var secret string
	var duration time.Duration

	if tokenType == AccessToken {
		secret = s.config.GetAccessTokenSecret()
		duration = s.config.GetAccessTokenDuration()
	} else {
		secret = s.config.GetRefreshTokenSecret()
		duration = s.config.GetRefreshTokenDuration()
	}

	claims := &Claims{
		UserID:    user.ID,
		SessionID: user.Token,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.GetAccessTokenSecret())
}

func (s *Service) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.GetRefreshTokenSecret())
}

func (s *Service) validateToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
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
	savedToken, err := s.tokenRepo.GetToken(context.Background(), claims.UserID)
	if err != nil {
		return nil, ErrTokenRevoked
	}

	// If no token found in Redis or the tokens don't match
	if savedToken != tokenString {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}
