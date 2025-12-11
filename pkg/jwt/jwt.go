package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Username  string `json:"username"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	accessTokenSecret    string
	refreshTokenSecret   string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTManager creates a new instance of JWTManager.
//
// It takes in the secrets for generating access and refresh tokens,
// as well as the duration for both tokens. It returns a pointer to
// the newly created JWTManager.
func NewJWTManager(accessSecret, refreshSecret string, accessDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		accessTokenSecret:    accessSecret,
		refreshTokenSecret:   refreshSecret,
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

func (m *JWTManager) GenerateTokenPair(userID, sessionID, role, username string) (string, string, error) {
	accessToken, err := m.generateToken(userID, sessionID, role, username, m.accessTokenSecret, m.accessTokenDuration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := m.generateToken(userID, sessionID, role, username, m.refreshTokenSecret, m.refreshTokenDuration)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (m *JWTManager) generateToken(userID, sessionID, role, username, secret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        sessionID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.accessTokenSecret)
}

func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return m.validateToken(tokenString, m.refreshTokenSecret)
}

func (m *JWTManager) validateToken(tokenString, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (m *JWTManager) GetRefreshTokenDuration() time.Duration {
	return m.refreshTokenDuration
}

func (m *JWTManager) GetAccessTokenDuration() time.Duration {
	return m.accessTokenDuration
}

func GenerateTestToken(userID, sessionID, role, username, secret string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID:    userID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
