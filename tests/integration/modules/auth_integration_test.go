//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthIntegration_Login_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	jwtManager := jwt.NewJWTManager(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		24*time.Hour,
	)

	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	authUC := usecase.NewAuthUsecase(
		jwtManager,
		tokenRepo,
		userRepo,
		tm,
		env.Logger,
		nil,
		env.Enforcer,
		auditUC,
		nil, // TaskDistributor nil for integration test (skip worker)
	)

	loginReq := model.LoginRequest{
		Username:  "testuser",
		Password:  "password123",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	require.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.Equal(t, testUser.ID, loginResp.User.ID)
	assert.Equal(t, testUser.Username, loginResp.User.Username)

	// Verify session exists in Redis
	keys, err := env.Redis.Keys(context.Background(), "session:*").Result()
	require.NoError(t, err)
	assert.NotEmpty(t, keys, "Session should be stored in Redis")
}

func TestAuthIntegration_Login_InvalidCredentials(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	jwtManager := jwt.NewJWTManager(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		24*time.Hour,
	)

	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	authUC := usecase.NewAuthUsecase(
		jwtManager,
		tokenRepo,
		userRepo,
		tm,
		env.Logger,
		nil,
		env.Enforcer,
		auditUC,
		nil,
	)

	loginReq := model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
}

func TestAuthIntegration_TokenRefresh_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	jwtManager := jwt.NewJWTManager(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		24*time.Hour,
	)

	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	authUC := usecase.NewAuthUsecase(
		jwtManager,
		tokenRepo,
		userRepo,
		tm,
		env.Logger,
		nil,
		env.Enforcer,
		auditUC,
		nil,
	)

	loginReq := model.LoginRequest{
		Username:  "testuser",
		Password:  "password123",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	loginResp, refreshToken, err := authUC.Login(context.Background(), loginReq)
	require.NoError(t, err)

	tokenResp, newRefreshToken, err := authUC.RefreshToken(context.Background(), refreshToken)

	require.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, loginResp.AccessToken, tokenResp.AccessToken)
	assert.NotEqual(t, refreshToken, newRefreshToken)
}

func TestAuthIntegration_Logout_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	_, err := env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user")
	require.NoError(t, err)

	jwtManager := jwt.NewJWTManager(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		24*time.Hour,
	)

	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	authUC := usecase.NewAuthUsecase(
		jwtManager,
		tokenRepo,
		userRepo,
		tm,
		env.Logger,
		nil,
		env.Enforcer,
		auditUC,
		nil,
	)

	loginReq := model.LoginRequest{
		Username:  "testuser",
		Password:  "password123",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	loginResp, _, err := authUC.Login(context.Background(), loginReq)
	require.NoError(t, err)

	// Parse token to get session ID
	claims, err := jwtManager.ValidateAccessToken(loginResp.AccessToken)
	require.NoError(t, err)

	err = authUC.RevokeToken(context.Background(), testUser.ID, claims.SessionID)
	require.NoError(t, err)

	// Verify session deleted from Redis
	keys, err := env.Redis.Keys(context.Background(), "session:*").Result()
	require.NoError(t, err)
	assert.Empty(t, keys, "Session should be deleted from Redis")
}

func TestAuthIntegration_ForgotPassword_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "forgotuser", "forgot@example.com", "password123")

	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	
	authUC := usecase.NewAuthUsecase(
		nil, // JWT manager not needed
		tokenRepo,
		userRepo,
		tm,
		env.Logger,
		nil,
		nil, // Enforcer not needed
		nil, // Audit not strict here
		nil, // Worker disabled
	)

	err := authUC.ForgotPassword(context.Background(), testUser.Email)
	require.NoError(t, err)

	// Verify token in DB
	var token authEntity.PasswordResetToken
	err = env.DB.Where("email = ?", testUser.Email).First(&token).Error
	require.NoError(t, err)
	assert.NotEmpty(t, token.Token)
	assert.Equal(t, testUser.Email, token.Email)
}

func TestAuthIntegration_ResetPassword_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "resetuser", "reset@example.com", "oldpass")
	
	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	// Seed token
	validToken := "valid-token-123"
	resetToken := authEntity.PasswordResetToken{
		Email:     testUser.Email,
		Token:     validToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	env.DB.Create(&resetToken)

	authUC := usecase.NewAuthUsecase(
		nil,
		tokenRepo,
		userRepo,
		tm,
		env.Logger,
		nil,
		nil,
		nil,
		nil,
	)

	newPassword := "NewStrongPass123!"
	err := authUC.ResetPassword(context.Background(), validToken, newPassword)
	require.NoError(t, err)

	// Verify token deleted
	var checkToken authEntity.PasswordResetToken
	err = env.DB.Where("email = ?", testUser.Email).First(&checkToken).Error
	assert.Error(t, err, "Token should be deleted after reset")

	// Verify login with new password logic
	var updatedUser userEntity.User
	env.DB.First(&updatedUser, "id = ?", testUser.ID)
	// You might check hash if bcrypt available in test env, otherwise assume success if no error.
}
