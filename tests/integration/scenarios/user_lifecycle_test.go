//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLifecycle_FullFlow(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	// Setup Dependencies
	jwtManager := jwt.NewJWTManager("lifecycle-secret", "lifecycle-refresh", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)

	authUC := authUseCase.NewAuthUsecase(jwtManager, tokenRepo, userRepo, tm, env.Logger, nil, env.Enforcer, auditUC, nil)
	userUC := userUseCase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC)

	ctx := context.Background()

	// 1. Registration
	regReq := &userModel.RegisterUserRequest{
		Username: "lifecycle", Email: "lifecycle@example.com", Password: "password123", Name: "Life Cycle",
	}
	_, err := userUC.Create(ctx, regReq)
	require.NoError(t, err)

	// 2. Login
	loginReq := authModel.LoginRequest{Username: regReq.Username, Password: regReq.Password}
	loginResp, _, err := authUC.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResp.AccessToken)

	// 3. Profile Update
	updateReq := &userModel.UpdateUserRequest{
		ID: loginResp.User.ID, Name: "Updated Life",
	}
	updateResp, err := userUC.Update(ctx, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Life", updateResp.Name)

	// 4. Delete
	deleteReq := &userModel.DeleteUserRequest{ID: loginResp.User.ID}
	err = userUC.DeleteUser(ctx, loginResp.User.ID, deleteReq)
	require.NoError(t, err)

	// 5. Verify Deletion
	_, err = userRepo.FindByID(ctx, loginResp.User.ID)
	assert.Error(t, err)
}
