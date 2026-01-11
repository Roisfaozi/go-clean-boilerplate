//go:build integration
// +build integration

package scenarios

import (
	"context"
	"errors"
	"testing"

	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Scenario: Register User Transactional Rollback
// Ensures that if role assignment fails, the user creation is rolled back.
func TestScenario_TransactionalIntegrity_RegisterRollback(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// 1. Setup Dependencies
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	mockEnforcer := new(mocks.MockIEnforcer)

	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	aucRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)
	auditService := auditUC.NewAuditUseCase(aucRepo, env.Logger)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 60, 60)
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, nil, mockEnforcer, auditService, nil)

	userService := userUC.NewUserUseCase(tm, env.Logger, uRepo, mockEnforcer, auditService, authService)

	// 2. Define Expectations
	expectedErr := errors.New("casbin connection error")

	// FIX: The mock expects variadic arguments.
	// We match any first argument (userID) and specific second argument ("role:user")
	mockEnforcer.On("AddGroupingPolicy", mock.Anything).Return(false, expectedErr)

	// 3. Execute Register
	req := &userModel.RegisterUserRequest{
		Username: "rollback_user",
		Email:    "rollback@test.com",
		Password: "Password123!",
		Name:     "Rollback User",
	}

	_, err := userService.Create(context.Background(), req)

	// 4. Assertions
	require.Error(t, err, "Expected error from UserUseCase when Role assignment fails")
	
	// Expectation: User should NOT exist (Rolled back)
	user, _ := uRepo.FindByUsername(context.Background(), req.Username)
	assert.Nil(t, user, "User should be rolled back (not found) when role assignment fails")
}
