//go:build e2e
// +build e2e

package api

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthE2E_EmailVerificationFlow(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	client := server.Client
	email := "verify@example.com"
	username := "verifyuser"

	// 1. Register User
	registerReq := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": "securePass123",
		"fullname": "Verify User",
	}
	resp := client.POST("/api/v1/users/register", registerReq)
	assert.Equal(t, 201, resp.StatusCode)

	var registerResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	err := resp.JSON(&registerResult)
	require.NoError(t, err)
	userID := registerResult.Data.ID

	// 2. Login
	loginReq := map[string]interface{}{
		"username": username,
		"password": "securePass123",
	}
	resp = client.POST("/api/v1/auth/login", loginReq)
	assert.Equal(t, 200, resp.StatusCode)

	var loginResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&loginResult)
	require.NoError(t, err)
	accessToken := loginResult.Data.AccessToken

	// 3. Request Verification Email (authenticated)
	resp = client.POST("/api/v1/auth/resend-verification", nil, setup.WithAuth(accessToken))
	assert.Equal(t, 200, resp.StatusCode)

	// 4. Get Token from DB (Backdoor for testing)
	var verificationToken entity.EmailVerificationToken
	err = server.DB.Where("email = ?", email).First(&verificationToken).Error
	require.NoError(t, err)
	require.NotEmpty(t, verificationToken.Token)

	// 5. Verify Email
	verifyReq := map[string]interface{}{
		"token": verificationToken.Token,
	}
	resp = client.POST("/api/v1/auth/verify-email", verifyReq)
	assert.Equal(t, 200, resp.StatusCode)

	// 6. Check DB: email_verified_at should be set
	var user struct {
		EmailVerifiedAt *int64 `gorm:"column:email_verified_at"`
	}
	err = server.DB.Table("users").Where("id = ?", userID).First(&user).Error
	require.NoError(t, err)
	assert.NotNil(t, user.EmailVerifiedAt, "email_verified_at should be set after verification")

	// 7. Try to request verification again - should fail (already verified)
	resp = client.POST("/api/v1/auth/resend-verification", nil, setup.WithAuth(accessToken))
	assert.Equal(t, 400, resp.StatusCode) // ErrAlreadyVerified
}

func TestAuthE2E_VerifyEmail_InvalidToken(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	verifyReq := map[string]interface{}{
		"token": "completely-invalid-token",
	}
	resp := server.Client.POST("/api/v1/auth/verify-email", verifyReq)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestAuthE2E_ResendVerification_Unauthenticated(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	// Try to resend without auth
	resp := server.Client.POST("/api/v1/auth/resend-verification", nil)
	assert.Equal(t, 401, resp.StatusCode)
}
