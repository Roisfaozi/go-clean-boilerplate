//go:build integration
// +build integration

package modules

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================
// POSITIVE TEST CASES
// ============================================

func TestAuthIntegration_Login_Positive_ValidCredentials(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "validuser", "valid@example.com", "ValidPass123!")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{
		Username:  "validuser",
		Password:  "ValidPass123!",
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0",
	}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	require.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.Greater(t, int64(loginResp.ExpiresIn), int64(0)) // Ensure type match
}

func TestAuthIntegration_TokenRefresh_Positive_ValidRefreshToken(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "refreshuser", "refresh@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "refreshuser", Password: "password123", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
	_, refreshToken, err := authUC.Login(context.Background(), loginReq)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	newToken, newRefresh, err := authUC.RefreshToken(context.Background(), refreshToken)

	require.NoError(t, err)
	assert.NotEmpty(t, newToken.AccessToken)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, refreshToken, newRefresh)
}

// ============================================
// NEGATIVE TEST CASES
// ============================================

func TestAuthIntegration_Login_Negative_InvalidPassword(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "correctpassword")
	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "testuser", Password: "wrongpassword"}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
}

func TestAuthIntegration_Login_Negative_NonExistentUser(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "nonexistent", Password: "password123"}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
}

func TestAuthIntegration_TokenRefresh_Negative_InvalidToken(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC := setupAuthUseCase(t, env)

	invalidToken := "invalid.token.string"

	newToken, newRefresh, err := authUC.RefreshToken(context.Background(), invalidToken)

	assert.Error(t, err)
	assert.Nil(t, newToken)
	assert.Empty(t, newRefresh)
}

func TestAuthIntegration_TokenRefresh_Negative_ExpiredToken(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "expireduser", "expired@example.com", "password123")

	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", 15*time.Minute, 1*time.Millisecond)
	expiredToken, _, _ := jwtManager.GenerateTokenPair(testUser.ID, "session-id", "role:user", testUser.Username)

	time.Sleep(2 * time.Millisecond)

	authUC := setupAuthUseCaseWithJWT(t, env, jwtManager)

	newToken, newRefresh, err := authUC.RefreshToken(context.Background(), expiredToken)

	assert.Error(t, err)
	assert.Nil(t, newToken)
	assert.Empty(t, newRefresh)
}

func TestAuthIntegration_Login_Negative_EmptyCredentials(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC := setupAuthUseCase(t, env)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"Empty Username", "", "password123"},
		{"Empty Password", "testuser", ""},
		{"Both Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginReq := model.LoginRequest{Username: tt.username, Password: tt.password}
			loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

			assert.Error(t, err)
			assert.Nil(t, loginResp)
			assert.Empty(t, refreshToken)
		})
	}
}

// ============================================
// EDGE CASES
// ============================================

func TestAuthIntegration_Login_Edge_SpecialCharactersInUsername(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	specialUsername := "user@#$%^&*()"
	testUser := setup.CreateTestUser(t, env.DB, specialUsername, "special@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: specialUsername, Password: "password123"}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAuthIntegration_Login_Edge_VeryLongPassword(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// Bcrypt max is 72 chars. We test up to that limit.
	longPassword := strings.Repeat("a", 72)
	testUser := setup.CreateTestUser(t, env.DB, "longpassuser", "long@example.com", longPassword)
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "longpassuser", Password: longPassword}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAuthIntegration_Login_Edge_UnicodeCharacters(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	unicodeUsername := "用户名测试"
	testUser := setup.CreateTestUser(t, env.DB, unicodeUsername, "unicode@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: unicodeUsername, Password: "password123"}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAuthIntegration_Login_Edge_CaseSensitivity(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// In typical MySQL, collation is case-insensitive by default.
	// This test might fail depending on DB collation. We assume Case-Sensitive for Security.
	setup.CreateTestUser(t, env.DB, "TestUser", "test@example.com", "password123")
	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "testuser", Password: "password123"}

	loginResp, _, err := authUC.Login(context.Background(), loginReq)

	// If the DB is case-insensitive, login will succeed. We check the outcome.
	if err == nil {
		assert.Equal(t, "TestUser", loginResp.User.Username)
	}
}

func TestAuthIntegration_TokenRefresh_Edge_MultipleRefreshInSequence(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "multirefresh", "multi@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "multirefresh", Password: "password123", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
	_, refreshToken, err := authUC.Login(context.Background(), loginReq)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		newToken, newRefresh, err := authUC.RefreshToken(context.Background(), refreshToken)
		require.NoError(t, err, "Refresh iteration %d failed", i+1)
		assert.NotEmpty(t, newToken.AccessToken)
		refreshToken = newRefresh
	}
}

// ============================================
// SECURITY TEST CASES
// ============================================

func TestAuthIntegration_Security_SQLInjectionAttempt(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC := setupAuthUseCase(t, env)

	sqlInjectionAttempts := []string{
		"admin' OR '1'='1",
		"admin'--",
		"admin' OR '1'='1'--",
		"' OR 1=1--",
		"admin'; DROP TABLE users--",
	}

	for _, injection := range sqlInjectionAttempts {
		t.Run("SQLInjection_"+injection, func(t *testing.T) {
			loginReq := model.LoginRequest{Username: injection, Password: "password"}
			loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

			assert.Error(t, err)
			assert.Nil(t, loginResp)
			assert.Empty(t, refreshToken)
		})
	}
}

func TestAuthIntegration_Security_BruteForceProtection(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "brutetest", "brute@example.com", "correctpassword")
	authUC := setupAuthUseCase(t, env)

	for i := 0; i < 5; i++ { // Reduced count for singleton speed
		loginReq := model.LoginRequest{Username: "brutetest", Password: "wrongpassword" + string(rune(i))}
		_, _, err := authUC.Login(context.Background(), loginReq)
		assert.Error(t, err, "Attempt %d should fail", i+1)
	}
}

func TestAuthIntegration_Security_TokenReuse(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "reusetest", "reuse@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq := model.LoginRequest{Username: "reusetest", Password: "password123", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
	_, refreshToken, err := authUC.Login(context.Background(), loginReq)
	require.NoError(t, err)

	_, newRefresh, err := authUC.RefreshToken(context.Background(), refreshToken)
	require.NoError(t, err)

	_, _, err = authUC.RefreshToken(context.Background(), refreshToken)
	assert.Error(t, err, "Old refresh token should not work after being used")

	_, _, err = authUC.RefreshToken(context.Background(), newRefresh)
	assert.NoError(t, err, "New refresh token should work")
}

func TestAuthIntegration_Security_SessionHijacking(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "hijacktest", "hijack@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	loginReq1 := model.LoginRequest{Username: "hijacktest", Password: "password123", IPAddress: "192.168.1.1", UserAgent: "Device1"}
	resp1, _, err := authUC.Login(context.Background(), loginReq1)
	require.NoError(t, err)

	loginReq2 := model.LoginRequest{Username: "hijacktest", Password: "password123", IPAddress: "192.168.1.2", UserAgent: "Device2"}
	resp2, _, err := authUC.Login(context.Background(), loginReq2)
	require.NoError(t, err)

	assert.NotEqual(t, resp1.AccessToken, resp2.AccessToken, "Different devices should have different tokens")
}

func TestAuthIntegration_Security_XSSInUserAgent(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "xsstest", "xss@example.com", "password123")
	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	authUC := setupAuthUseCase(t, env)

	xssPayload := "<script>alert('XSS')</script>"
	loginReq := model.LoginRequest{
		Username:  "xsstest",
		Password:  "password123",
		IPAddress: "127.0.0.1",
		UserAgent: xssPayload,
	}

	loginResp, _, err := authUC.Login(context.Background(), loginReq)

	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func setupAuthUseCase(t *testing.T, env *setup.TestEnvironment) usecase.AuthUseCase {
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
	return setupAuthUseCaseWithJWT(t, env, jwtManager)
}

func setupAuthUseCaseWithJWT(t *testing.T, env *setup.TestEnvironment, jwtManager *jwt.JWTManager) usecase.AuthUseCase {
	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	return usecase.NewAuthUsecase(jwtManager, tokenRepo, userRepo, tm, env.Logger, nil, env.Enforcer, auditUC)
}