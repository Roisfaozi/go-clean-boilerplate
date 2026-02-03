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
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_Auth_ConcurrentSessions(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Logger)
	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, oRepo, tm, env.Logger, nil, nil, env.Enforcer, nil, nil)

	password := "Pass123!"
	user := setup.CreateTestUser(t, env.DB, "multi_session_user", "multi@test.com", password)

	loginA, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: password, UserAgent: "Browser A",
	})
	require.NoError(t, err)
	tokenA := loginA.AccessToken

	loginB, _, err := authService.Login(context.Background(), authModel.LoginRequest{
		Username: user.Username, Password: password, UserAgent: "Browser B",
	})
	require.NoError(t, err)
	tokenB := loginB.AccessToken

	sessions, err := authService.GetUserSessions(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, sessions, 2, "User should have 2 active sessions")

	claimsA, _ := jwtManager.ValidateAccessToken(tokenA)
	err = authService.RevokeToken(context.Background(), user.ID, claimsA.SessionID)
	require.NoError(t, err)

	_, err = authService.ValidateAccessToken(tokenA)
	assert.Error(t, err, "Session A should be revoked")

	claimsB, err := authService.ValidateAccessToken(tokenB)
	assert.NoError(t, err, "Session B should remain active")
	assert.Equal(t, user.ID, claimsB.UserID)

	sessionsAfter, _ := authService.GetUserSessions(context.Background(), user.ID)
	assert.Len(t, sessionsAfter, 1)
	assert.Equal(t, claimsB.SessionID, sessionsAfter[0].ID)
}
