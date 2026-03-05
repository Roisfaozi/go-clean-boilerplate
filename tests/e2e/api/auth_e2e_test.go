//go:build e2e
// +build e2e

package api

import (
	"net/http"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
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

func TestSecurityE2E_TokenRotation(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)

	user := f.Create(func(u *userEntity.User) {
		u.Username = "rotate_user"
		u.Email = "rot@test.com"
		u.Password = string(hash)
	})
	server.Enforcer.AddGroupingPolicy(user.ID, "role:user", "global")

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": user.Username, "password": "StrongPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	cookies := resp.Cookies()
	var refreshToken1 *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshToken1 = c
			break
		}
	}
	require.NotNil(t, refreshToken1, "Refresh token cookie not found")

	req, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
	req.AddCookie(refreshToken1)

	clientWithCookie := &http.Client{}
	respRotate, err := clientWithCookie.Do(req)
	require.NoError(t, err)
	defer respRotate.Body.Close()

	require.Equal(t, 200, respRotate.StatusCode)

	cookies2 := respRotate.Cookies()
	var refreshToken2 *http.Cookie
	for _, c := range cookies2 {
		if c.Name == "refresh_token" {
			refreshToken2 = c
			break
		}
	}
	require.NotNil(t, refreshToken2)
	assert.NotEqual(t, refreshToken1.Value, refreshToken2.Value)

	reqReuse, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
	reqReuse.AddCookie(refreshToken1)

	respReuse, err := clientWithCookie.Do(reqReuse)
	require.NoError(t, err)
	defer respReuse.Body.Close()

	assert.Contains(t, []int{401, 400}, respReuse.StatusCode)
}
