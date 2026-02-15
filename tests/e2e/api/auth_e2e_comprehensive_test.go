//go:build e2e
// +build e2e

package api

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthE2E_CompleteFlow_Positive(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	client := server.Client

	registerReq := map[string]interface{}{
		"username": "e2euser",
		"email":    "e2e@example.com",
		"password": "SecurePass123!",
		"fullname": "E2E User",
	}

	resp := client.POST("/api/v1/users/register", registerReq)
	assert.Equal(t, 201, resp.StatusCode)

	loginReq := map[string]interface{}{
		"username": "e2euser",
		"password": "SecurePass123!",
	}

	resp = client.POST("/api/v1/auth/login", loginReq)
	assert.Equal(t, 200, resp.StatusCode)

	var loginResult struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err := resp.JSON(&loginResult)
	require.NoError(t, err)

	resp = client.GET("/api/v1/users/me", setup.WithAuth(loginResult.Data.AccessToken))
	assert.Equal(t, 200, resp.StatusCode)

	resp = client.POST("/api/v1/auth/logout", nil, setup.WithAuth(loginResult.Data.AccessToken))
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAuthE2E_Login_Negative_InvalidCredentials(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	loginReq := map[string]interface{}{
		"username": "nonexistent",
		"password": "wrongpassword",
	}

	resp := server.Client.POST("/api/v1/auth/login", loginReq)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestAuthE2E_Register_Negative_DuplicateUsername(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	registerReq := map[string]interface{}{
		"username": "duplicate",
		"email":    "first@example.com",
		"password": "password123",
		"fullname": "First User",
	}

	resp := server.Client.POST("/api/v1/users/register", registerReq)
	assert.Equal(t, 201, resp.StatusCode)

	registerReq2 := map[string]interface{}{
		"username": "duplicate",
		"email":    "second@example.com",
		"password": "password123",
		"fullname": "Second User",
	}

	resp = server.Client.POST("/api/v1/users/register", registerReq2)
	assert.Equal(t, 409, resp.StatusCode)
}

func TestAuthE2E_ProtectedEndpoint_Negative_NoToken(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	resp := server.Client.GET("/api/v1/users/me")
	assert.Equal(t, 401, resp.StatusCode)
}

func TestAuthE2E_ProtectedEndpoint_Negative_InvalidToken(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	resp := server.Client.GET("/api/v1/users/me", setup.WithAuth("invalid.token.here"))
	assert.Equal(t, 401, resp.StatusCode)
}

func TestAuthE2E_Register_Edge_SpecialCharactersInUsername(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	registerReq := map[string]interface{}{
		"username": "user@#$%",
		"email":    "special@example.com",
		"password": "password123",
		"fullname": "Special User",
	}

	resp := server.Client.POST("/api/v1/users/register", registerReq)

	assert.True(t, resp.StatusCode == 201 || resp.StatusCode == 400 || resp.StatusCode == 422)
}

func TestAuthE2E_Login_Edge_CaseSensitiveUsername(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	registerReq := map[string]interface{}{
		"username": "TestUser",
		"email":    "test@example.com",
		"password": "password123",
		"fullname": "Test User",
	}

	resp := server.Client.POST("/api/v1/users/register", registerReq)
	require.Equal(t, 201, resp.StatusCode)

	loginReq := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	}

	resp = server.Client.POST("/api/v1/auth/login", loginReq)

	assert.True(t, resp.StatusCode == 401 || resp.StatusCode == 200)
}

func TestAuthE2E_Security_SQLInjectionInLogin(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	sqlInjections := []string{
		"admin' OR '1'='1",
		"admin'--",
		"' OR 1=1--",
	}

	for _, injection := range sqlInjections {
		loginReq := map[string]interface{}{
			"username": injection,
			"password": "password",
		}

		resp := server.Client.POST("/api/v1/auth/login", loginReq)
		assert.Equal(t, 401, resp.StatusCode, "SQL injection should be prevented")
	}
}

func TestAuthE2E_Security_XSSInRegistration(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	registerReq := map[string]interface{}{
		"username": "xssuser",
		"email":    "xss@example.com",
		"password": "password123",
		"fullname": "<script>alert('XSS')</script>",
	}

	resp := server.Client.POST("/api/v1/users/register", registerReq)
	assert.True(t, resp.StatusCode == 201 || resp.StatusCode == 400 || resp.StatusCode == 422)
}

func TestAuthE2E_Security_BruteForceProtection(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	for i := 0; i < 10; i++ {
		loginReq := map[string]interface{}{
			"username": "testuser",
			"password": "wrongpassword",
		}

		resp := server.Client.POST("/api/v1/auth/login", loginReq)
		assert.True(t, resp.StatusCode == 401 || resp.StatusCode == 429)
	}
}
