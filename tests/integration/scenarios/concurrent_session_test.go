//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Scenario: Concurrent Multi-Session Management
// Verifies that a user can have multiple active sessions and revoking one doesn't affect others.
func TestScenario_Auth_ConcurrentSessions(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// 1. Setup Modules
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, nil, env.Enforcer, nil, nil)

	// 2. Create User
	password := "Pass123!"
	user := setup.CreateTestUser(t, env.DB, "multi_session_user", "multi@test.com", password)

	// 3. Login from "Browser A"
	loginA, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: password, UserAgent: "Browser A",
	})
	require.NoError(t, err)
	tokenA := loginA.AccessToken

	// 4. Login from "Browser B"
	loginB, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: password, UserAgent: "Browser B",
	})
	require.NoError(t, err)
	tokenB := loginB.AccessToken

	// 5. Verify both sessions are active in Redis
	sessions, err := authService.GetUserSessions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, sessions, 2, "User should have 2 active sessions")

	// 6. Logout from "Browser A"
	claimsA, _ := jwtManager.ValidateAccessToken(tokenA)
	err = authService.RevokeToken(context.Background(), user.ID, claimsA.SessionID)
	require.NoError(t, err)

	// 7. Verify "Browser A" token is invalid
	_, err = authService.ValidateAccessToken(tokenA)
	assert.Error(t, err, "Session A should be revoked")

	// 8. Verify "Browser B" token is STILL VALID
	claimsB, err := authService.ValidateAccessToken(tokenB)
	assert.NoError(t, err, "Session B should remain active")
	assert.Equal(t, user.ID, claimsB.UserID)

	// 9. Verify only 1 session remains in Redis
	sessionsAfter, _ := authService.GetUserSessions(context.Background(), user.ID)
	assert.Len(t, sessionsAfter, 1)
	assert.Equal(t, claimsB.SessionID, sessionsAfter[0].ID)
}
