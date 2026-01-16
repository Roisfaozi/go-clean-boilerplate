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

func TestAuthE2E_RegisterLoginLogout(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	client := server.Client

	// Register
	registerReq := map[string]interface{}{
		"username": "e2euser",
		"email":    "e2e@example.com",
		"password": "password123",
		"fullname": "E2E User",
	}

	resp := client.POST("/api/v1/users/register", registerReq)
	assert.Equal(t, 201, resp.StatusCode)

	var registerResult struct {
		Data struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"data"`
	}
	err := resp.JSON(&registerResult)
	require.NoError(t, err)
	assert.Equal(t, "e2euser", registerResult.Data.Username)

	// Login
	loginReq := map[string]interface{}{
		"username": "e2euser",
		"password": "password123",
	}

	resp = client.POST("/api/v1/auth/login", loginReq)
	assert.Equal(t, 200, resp.StatusCode)

	var loginResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
		} `json:"data"`
	}
	err = resp.JSON(&loginResult)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResult.Data.AccessToken)

	// Access Protected Endpoint
	resp = client.GET("/api/v1/users/me", setup.WithAuth(loginResult.Data.AccessToken))
	assert.Equal(t, 200, resp.StatusCode)

	// Logout
	resp = client.POST("/api/v1/auth/logout", nil, setup.WithAuth(loginResult.Data.AccessToken))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAuthE2E_InvalidCredentials(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	loginReq := map[string]interface{}{
		"username": "nonexistent",
		"password": "wrongpassword",
	}

	resp := server.Client.POST("/api/v1/auth/login", loginReq)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestAuthE2E_ForgotPasswordFlow(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	client := server.Client
	email := "recovery@example.com"
	username := "recoveryuser"

	// 1. Register User
	registerReq := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": "oldPassword123",
		"fullname": "Recovery User",
	}
	resp := client.POST("/api/v1/users/register", registerReq)
	assert.Equal(t, 201, resp.StatusCode)

	// 2. Request Forgot Password
	forgotReq := map[string]interface{}{"email": email}
	resp = client.POST("/api/v1/auth/forgot-password", forgotReq)
	assert.Equal(t, 200, resp.StatusCode)

	// 3. Get Token from DB (Backdoor for testing only)
	var resetToken entity.PasswordResetToken
	err := server.DB.Where("email = ?", email).First(&resetToken).Error
	require.NoError(t, err)
	require.NotEmpty(t, resetToken.Token)

	// 4. Reset Password
	newPassword := "brandNewPass2026!"
	resetReq := map[string]interface{}{
		"token":        resetToken.Token,
		"new_password": newPassword,
	}
	resp = client.POST("/api/v1/auth/reset-password", resetReq)
	assert.Equal(t, 200, resp.StatusCode)

	// 5. Login with NEW password
	loginReq := map[string]interface{}{
		"username": username,
		"password": newPassword,
	}
	resp = client.POST("/api/v1/auth/login", loginReq)
	assert.Equal(t, 200, resp.StatusCode)

	// 6. Login with OLD password should FAIL
	oldLoginReq := map[string]interface{}{
		"username": username,
		"password": "oldPassword123",
	}
	resp = client.POST("/api/v1/auth/login", oldLoginReq)
	assert.Equal(t, 401, resp.StatusCode)
}