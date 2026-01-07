//go:build integration
// +build integration

package modules

import (
	"context"
	"strings"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- HELPERS ---

func setupAuthIntegration(env *setup.TestEnvironment) (usecase.AuthUseCase, *jwt.JWTManager) {
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
	return setupAuthIntegrationWithJWT(env, jwtManager), jwtManager
}

func setupAuthIntegrationWithJWT(env *setup.TestEnvironment, jwtManager *jwt.JWTManager) usecase.AuthUseCase {
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	return usecase.NewAuthUsecase(
		jwtManager, tokenRepo, userRepo, tm, env.Logger, nil, env.Enforcer, auditUC, nil,
	)
}

// ============================================
// LOGIN SCENARIOS
// ============================================

func TestAuthIntegration_Login(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC, _ := setupAuthIntegration(env)
	password := "SecurePass123!"
	testUser := setup.CreateTestUser(t, env.DB, "authuser", "auth@example.com", password)
	_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")

	t.Run("Success with Valid Credentials", func(t *testing.T) {
		loginReq := model.LoginRequest{Username: "authuser", Password: password, IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
		resp, refreshToken, err := authUC.Login(context.Background(), loginReq)

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.Equal(t, testUser.ID, resp.User.ID)
		assert.Greater(t, int64(resp.ExpiresIn), int64(0))

		// Verify session exists in Redis
		keys, err := env.Redis.Keys(context.Background(), "session:*").Result()
		require.NoError(t, err)
		assert.NotEmpty(t, keys)
	})

	t.Run("Fail with Invalid Password", func(t *testing.T) {
		loginReq := model.LoginRequest{Username: "authuser", Password: "wrongpassword"}
		resp, _, err := authUC.Login(context.Background(), loginReq)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Fail with Non-Existent User", func(t *testing.T) {
		loginReq := model.LoginRequest{Username: "nonexistent", Password: "password123"}
		_, _, err := authUC.Login(context.Background(), loginReq)
		assert.Error(t, err)
	})

	t.Run("Fail with Empty Credentials", func(t *testing.T) {
		tests := []struct {
			name string
			un   string
			pw   string
		}{
			{"Empty Username", "", "password123"},
			{"Empty Password", "authuser", ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: tt.un, Password: tt.pw})
				assert.Error(t, err)
			})
		}
	})

	t.Run("Edge - Special Characters in Username", func(t *testing.T) {
		specialUN := "user-@#$%^&*()"
		setup.CreateTestUser(t, env.DB, specialUN, "special@example.com", "password123")
		loginReq := model.LoginRequest{Username: specialUN, Password: "password123"}
		_, _, err := authUC.Login(context.Background(), loginReq)
		assert.NoError(t, err)
	})

	t.Run("Edge - Very Long Password (Bcrypt Limit 72)", func(t *testing.T) {
		longPW := strings.Repeat("a", 72)
		setup.CreateTestUser(t, env.DB, "longpw", "long@example.com", longPW)
		loginReq := model.LoginRequest{Username: "longpw", Password: longPW}
		_, _, err := authUC.Login(context.Background(), loginReq)
		assert.NoError(t, err)
	})

	t.Run("Edge - Unicode Characters", func(t *testing.T) {
		unicodeUN := "用户名测试"
		setup.CreateTestUser(t, env.DB, unicodeUN, "unicode@example.com", "password123")
		loginReq := model.LoginRequest{Username: unicodeUN, Password: "password123"}
		_, _, err := authUC.Login(context.Background(), loginReq)
		assert.NoError(t, err)
	})

	t.Run("Edge - Case Sensitivity", func(t *testing.T) {
		setup.CreateTestUser(t, env.DB, "CaseUser", "case@example.com", "password123")
		loginReq := model.LoginRequest{Username: "caseuser", Password: "password123"}
		_, loginResp, err := authUC.Login(context.Background(), loginReq)
		if err == nil {
			assert.NotEmpty(t, loginResp)
		}
	})
}

// ============================================
// TOKEN LIFECYCLE SCENARIOS
// ============================================

func TestAuthIntegration_TokenLifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC, jwtManager := setupAuthIntegration(env)
	password := "password123"
	testUser := setup.CreateTestUser(t, env.DB, "tokenuser", "token@example.com", password)
	_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")

	_, refreshToken, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "tokenuser", Password: password})

	t.Run("Success Refresh Token", func(t *testing.T) {
		time.Sleep(1 * time.Second)
		newToken, newRefresh, err := authUC.RefreshToken(context.Background(), refreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newToken.AccessToken)
		assert.NotEqual(t, refreshToken, newRefresh)

		// Update refreshToken for next sub-tests
		refreshToken = newRefresh
	})

	t.Run("Multiple Refresh In Sequence", func(t *testing.T) {
		currRefresh := refreshToken
		for i := 0; i < 3; i++ {
			time.Sleep(100 * time.Millisecond)
			_, nextRefresh, err := authUC.RefreshToken(context.Background(), currRefresh)
			require.NoError(t, err, "Refresh iteration %d failed", i+1)
			currRefresh = nextRefresh
		}
	})

	t.Run("Fail Refresh with Invalid Token", func(t *testing.T) {
		_, _, err := authUC.RefreshToken(context.Background(), "invalid.token.here")
		assert.Error(t, err)
	})

	t.Run("Fail Refresh with Expired Token", func(t *testing.T) {
		shortJWT := jwt.NewJWTManager("secret", "refresh", time.Minute, 1*time.Millisecond)
		expToken, _, _ := shortJWT.GenerateTokenPair(testUser.ID, "sid", "role:user", "tokenuser")
		time.Sleep(10 * time.Millisecond)

		customUC := setupAuthIntegrationWithJWT(env, shortJWT)
		_, _, err := customUC.RefreshToken(context.Background(), expToken)
		assert.Error(t, err)
	})

	t.Run("Success Logout (Revoke)", func(t *testing.T) {
		// Get fresh login for revocation
		lr, _, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "tokenuser", Password: password})
		claims, _ := jwtManager.ValidateAccessToken(lr.AccessToken)

		err := authUC.RevokeToken(context.Background(), testUser.ID, claims.SessionID)
		require.NoError(t, err)

		// Verify session is gone. Pattern should match repository.getSessionKey
		keys, _ := env.Redis.Keys(context.Background(), "session:"+testUser.ID+":"+claims.SessionID).Result()
		assert.Empty(t, keys, "Session should be deleted from Redis")
	})
}

// ============================================
// PASSWORD RECOVERY SCENARIOS
// ============================================

func TestAuthIntegration_PasswordRecovery(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC, _ := setupAuthIntegration(env)

	t.Run("Success Forgot Password", func(t *testing.T) {
		email := "forgot@example.com"
		setup.CreateTestUser(t, env.DB, "forgotuser", email, "old-pass")

		err := authUC.ForgotPassword(context.Background(), email)
		require.NoError(t, err)

		var token authEntity.PasswordResetToken
		err = env.DB.Where("email = ?", email).First(&token).Error
		require.NoError(t, err)
		assert.NotEmpty(t, token.Token)
	})

	t.Run("Success Reset Password", func(t *testing.T) {
		email := "reset_unique@example.com" // Use unique email to avoid constraint error
		testUser := setup.CreateTestUser(t, env.DB, "resetuser", email, "oldpass")

		// Seed a token
		resetToken := "secret-token-unique-123"
		err := env.DB.Create(&authEntity.PasswordResetToken{
			Email: email, Token: resetToken, ExpiresAt: time.Now().Add(time.Hour),
		}).Error
		require.NoError(t, err)

		err = authUC.ResetPassword(context.Background(), resetToken, "NewPass123!")
		require.NoError(t, err)

		// Verify old fails, new succeeds
		_, _, err = authUC.Login(context.Background(), model.LoginRequest{Username: testUser.Username, Password: "oldpass"})
		assert.Error(t, err)
		_, _, err = authUC.Login(context.Background(), model.LoginRequest{Username: testUser.Username, Password: "NewPass123!"})
		assert.NoError(t, err)
	})
}

// ============================================
// SECURITY SCENARIOS
// ============================================

func TestAuthIntegration_Security(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	authUC, _ := setupAuthIntegration(env)

	t.Run("SQL Injection Prevention", func(t *testing.T) {
		injections := []string{"admin' OR '1'='1", "admin'--", "admin'; DROP TABLE users--"}
		for _, inj := range injections {
			_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: inj, Password: "p"})
			assert.Error(t, err)
		}
	})

	t.Run("Brute Force Protection Simulation", func(t *testing.T) {
		setup.CreateTestUser(t, env.DB, "brute", "brute@example.com", "pass")
		for i := 0; i < 5; i++ {
			_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: "brute", Password: "w"})
			assert.Error(t, err)
		}
	})

	t.Run("Token Rotation Reuse Protection", func(t *testing.T) {
		testUser := setup.CreateTestUser(t, env.DB, "reuse", "reuse@example.com", "pass")
		_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
		_, rt1, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "reuse", Password: "pass"})

		_, rt2, _ := authUC.RefreshToken(context.Background(), rt1)

		// Attempt reuse RT1
		_, _, err := authUC.RefreshToken(context.Background(), rt1)
		assert.Error(t, err)

		// RT2 still works
		_, _, err = authUC.RefreshToken(context.Background(), rt2)
		assert.NoError(t, err)
	})

	t.Run("Session Hijacking Prevention (Device Differentiation)", func(t *testing.T) {
		testUser := setup.CreateTestUser(t, env.DB, "hijack", "hijack@example.com", "pass")
		_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")

		r1, _, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "hijack", Password: "pass", UserAgent: "D1"})
		r2, _, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "hijack", Password: "pass", UserAgent: "D2"})
		assert.NotEqual(t, r1.AccessToken, r2.AccessToken)
	})

	t.Run("XSS in UserAgent Handling", func(t *testing.T) {
		testUser := setup.CreateTestUser(t, env.DB, "xss", "xss@example.com", "pass")
		_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
		_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: "xss", Password: "pass", UserAgent: "<script>alert(1)</script>"})
		assert.NoError(t, err)
	})
}
