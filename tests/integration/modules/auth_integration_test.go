package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthIntegration(env *setup.TestEnvironment) usecase.AuthUseCase {
	tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	userRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	orgRepo := orgRepo.NewOrganizationRepository(env.DB)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	// Use TaskDistributor from env which is likely a mock or no-op in singleton env,
	// or we can use a real one if Redis is available.
	// For integration, we care about the flow, not actual email sending.
	// But we need to ensure DistributeTask* methods don't panic.
	// The env.TaskDistributor should handle this.

	// Authz adapter
	authz := repository.NewCasbinAdapter(env.Enforcer, "role:user", "global")

	jwtManager := jwt.NewJWTManager("secret", "refresh-secret", 15*time.Minute, 24*time.Hour)

	return usecase.NewAuthUsecase(
		5, // Max attempts
		15*time.Minute, // Lockout
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		env.Logger,
		nil, // Publisher (optional)
		authz,
		env.TaskDistributor,
		nil, // Ticket manager (optional)
	)
}

func TestAuthIntegration_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	authUC := setupAuthIntegration(env)
	ctx := context.Background()

	unique := fmt.Sprintf("user_%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s@example.com", unique)
	password := "StrongPass123!"

	// 1. Register
	regReq := model.RegisterRequest{
		Username: unique,
		Email:    email,
		Password: password,
		Name:     "Integration User",
	}
	loginRes, refreshToken, err := authUC.Register(ctx, regReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginRes.AccessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, unique, loginRes.User.Username)

	// 2. Validate Token
	claims, err := authUC.ValidateAccessToken(loginRes.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, loginRes.User.ID, claims.UserID)

	// 3. Refresh Token
	tokenRes, newRefresh, err := authUC.RefreshToken(ctx, refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenRes.AccessToken)
	assert.NotEmpty(t, newRefresh)
	assert.NotEqual(t, loginRes.AccessToken, tokenRes.AccessToken)

	// 4. Logout
	err = authUC.RevokeToken(ctx, claims.UserID, claims.SessionID)
	require.NoError(t, err)

	// 5. Verify Revocation
	_, err = authUC.ValidateAccessToken(loginRes.AccessToken)
	assert.ErrorIs(t, err, usecase.ErrTokenRevoked)
}

func TestAuthIntegration_AccountLockout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	authUC := setupAuthIntegration(env)
	ctx := context.Background()

	// Register user first
	unique := fmt.Sprintf("lockout_%d", time.Now().UnixNano())
	regReq := model.RegisterRequest{
		Username: unique,
		Email:    fmt.Sprintf("%s@example.com", unique),
		Password: "StrongPass123!",
		Name:     "Lockout User",
	}
	_, _, err := authUC.Register(ctx, regReq)
	require.NoError(t, err)

	// Attempt login with wrong password 5 times
	req := model.LoginRequest{Username: unique, Password: "WrongPassword"}
	for i := 0; i < 5; i++ {
		_, _, err := authUC.Login(ctx, req)
		assert.Error(t, err) // Invalid credentials
	}

	// 6th attempt should be locked
	_, _, err = authUC.Login(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many failed attempts")
}

func TestAuthIntegration_PasswordReset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	authUC := setupAuthIntegration(env)
	ctx := context.Background()

	unique := fmt.Sprintf("reset_%d", time.Now().UnixNano())
	regReq := model.RegisterRequest{
		Username: unique,
		Email:    fmt.Sprintf("%s@example.com", unique),
		Password: "OldPassword123!",
		Name:     "Reset User",
	}
	_, _, err := authUC.Register(ctx, regReq)
	require.NoError(t, err)

	// 1. Forgot Password (Generate Token)
	err = authUC.ForgotPassword(ctx, regReq.Email)
	require.NoError(t, err)

	// Inspect DB to get the token (since email is mocked/async)
	// We need direct DB access here or a way to intercept the token.
	// Since TokenRepo stores it in DB, we can query it.
	var tokenEntity struct {
		Token string
	}
	// Raw query to fetch token
	err = env.DB.Raw("SELECT token FROM password_reset_tokens WHERE email = ? ORDER BY expires_at DESC LIMIT 1", regReq.Email).Scan(&tokenEntity).Error
	require.NoError(t, err)
	require.NotEmpty(t, tokenEntity.Token)

	// 2. Reset Password
	newPass := "NewStrongPass123!"
	err = authUC.ResetPassword(ctx, tokenEntity.Token, newPass)
	require.NoError(t, err)

	// 3. Login with New Password
	loginReq := model.LoginRequest{Username: unique, Password: newPass}
	res, _, err := authUC.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)

	// 4. Login with Old Password (should fail)
	loginReqOld := model.LoginRequest{Username: unique, Password: "OldPassword123!"}
	_, _, err = authUC.Login(ctx, loginReqOld)
	assert.Error(t, err)
}

func TestAuthIntegration_SQLInjection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	authUC := setupAuthIntegration(env)
	ctx := context.Background()

	// Try SQL injection in login
	req := model.LoginRequest{
		Username: "' OR '1'='1",
		Password: "any",
	}
	_, _, err := authUC.Login(ctx, req)
	assert.Error(t, err)
	// Should not be 500
	assert.NotEqual(t, usecase.ErrInvalidCredentials, err)
}

func TestAuthIntegration_XSS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	authUC := setupAuthIntegration(env)
	ctx := context.Background()

	// XSS Payload in Name
	xssName := "<script>alert(1)</script>"
	regReq := model.RegisterRequest{
		Username: fmt.Sprintf("xss_%d", time.Now().UnixNano()),
		Email:    fmt.Sprintf("xss_%d@example.com", time.Now().UnixNano()),
		Password: "Password123!",
		Name:     xssName,
	}

	res, _, err := authUC.Register(ctx, regReq)
	require.NoError(t, err)

	// SanitizeString uses html.EscapeString
	expected := pkg.SanitizeString(xssName)
	assert.Equal(t, expected, res.User.Name)
	assert.False(t, strings.Contains(res.User.Name, "<script>"))
}
