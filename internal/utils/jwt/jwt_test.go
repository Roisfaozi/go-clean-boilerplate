package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testAccessSecret  = "test-access-secret-for-jwt"
	testRefreshSecret = "test-refresh-secret-for-jwt"
	testUserID        = "user-12345"
	testSessionID     = "session-67890"
)

func TestNewJWTManager(t *testing.T) {
	manager := NewJWTManager(testAccessSecret, testRefreshSecret, time.Hour, 24*time.Hour)
	assert.NotNil(t, manager)
	assert.Equal(t, testAccessSecret, manager.accessTokenSecret)
	assert.Equal(t, testRefreshSecret, manager.refreshTokenSecret)
	assert.Equal(t, time.Hour, manager.accessTokenDuration)
	assert.Equal(t, 24*time.Hour, manager.refreshTokenDuration)
}

func TestGenerateAndValidateTokenPair_Success(t *testing.T) {
	manager := NewJWTManager(testAccessSecret, testRefreshSecret, time.Minute*15, time.Hour*72)

	accessToken, refreshToken, err := manager.GenerateTokenPair(testUserID, testSessionID)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// Validate Access Token
	accessClaims, err := manager.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.NotNil(t, accessClaims)
	assert.Equal(t, testUserID, accessClaims.UserID)
	assert.Equal(t, testSessionID, accessClaims.SessionID)
	assert.Equal(t, testUserID, accessClaims.Subject)
	assert.WithinDuration(t, time.Now().Add(time.Minute*15), accessClaims.ExpiresAt.Time, time.Second*5)

	// Validate Refresh Token
	refreshClaims, err := manager.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotNil(t, refreshClaims)
	assert.Equal(t, testUserID, refreshClaims.UserID)
	assert.Equal(t, testSessionID, refreshClaims.SessionID)
	assert.WithinDuration(t, time.Now().Add(time.Hour*72), refreshClaims.ExpiresAt.Time, time.Second*5)
}

func TestValidateToken_Expired(t *testing.T) {
	// Create a manager that generates tokens that expire almost immediately
	manager := NewJWTManager(testAccessSecret, testRefreshSecret, -time.Second, -time.Second)

	accessToken, _, err := manager.GenerateTokenPair(testUserID, testSessionID)
	assert.NoError(t, err)

	// Allow a moment for the token to be definitively expired
	time.Sleep(10 * time.Millisecond)

	claims, err := manager.ValidateAccessToken(accessToken)
	assert.Error(t, err, "Validation should fail for an expired token")
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	manager1 := NewJWTManager("secret-one", "refresh-one", time.Minute, time.Hour)
	manager2 := NewJWTManager("secret-two", "refresh-two", time.Minute, time.Hour)

	// Generate token with manager1
	accessToken, _, err := manager1.GenerateTokenPair(testUserID, testSessionID)
	assert.NoError(t, err)

	// Try to validate with manager2 (which has a different secret)
	claims, err := manager2.ValidateAccessToken(accessToken)
	assert.Error(t, err, "Validation should fail for a token with a wrong signature")
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestValidateToken_MalformedToken(t *testing.T) {
	manager := NewJWTManager(testAccessSecret, testRefreshSecret, time.Minute, time.Hour)
	malformedToken := "this.is.not.a.valid.jwt"

	claims, err := manager.ValidateAccessToken(malformedToken)
	assert.Error(t, err, "Validation should fail for a malformed token")
	assert.Nil(t, claims)
}
