//go:build integration
// +build integration

package scenarios

import (
	"context"
	"errors"
	"testing"

	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
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

// TestScenario_TransactionalIntegrity_DeleteRollback verifies that
// if Audit Log fails during Delete User, the transaction is rolled back:
// 1. User is NOT deleted from DB.
// 2. User's Roles are NOT deleted from Casbin.
func TestScenario_TransactionalIntegrity_DeleteRollback(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	// 1. Setup Dependencies
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	
	// Real dependencies for setup
	realAuditRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)
	realAuditUC := auditUC.NewAuditUseCase(realAuditRepo, env.Logger)
	
	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 60, 60)
	authService := authUC.NewAuthUsecase(jwtManager, tRepo, uRepo, tm, env.Logger, nil, env.Enforcer, realAuditUC, nil)

	// Create User for test using real service
	setupService := userUC.NewUserUseCase(tm, env.Logger, uRepo, env.Enforcer, realAuditUC, authService)
	regReq := &userModel.RegisterUserRequest{
		Username: "todelete", Email: "delete@test.com", Password: "Pass123!", Name: "To Delete",
	}
	userResp, err := setupService.Create(context.Background(), regReq)
	require.NoError(t, err)

	// Verify User exists
	user, err := uRepo.FindByID(context.Background(), userResp.ID)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Verify Role assigned (default role:user)
	roles, err := env.Enforcer.GetRolesForUser(user.ID)
	require.NoError(t, err)
	require.Contains(t, roles, "role:user")

	// 2. Prepare Mock Audit for Failure
	mockAuditUC := new(mocks.MockAuditUseCase)
	mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("intentional audit failure"))

	// 3. Setup UserUseCase with MOCK Audit
	targetService := userUC.NewUserUseCase(tm, env.Logger, uRepo, env.Enforcer, mockAuditUC, authService)

	// 4. Execute Delete
	delReq := &userModel.DeleteUserRequest{ID: user.ID}
	err = targetService.DeleteUser(context.Background(), "admin-id", delReq)

	// 5. Assertions
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal server error") // Wrapped error

	// Verify User STILL exists (Rollback)
	userAfter, err := uRepo.FindByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, userAfter)
	assert.Equal(t, user.ID, userAfter.ID)

	// Verify Role STILL exists (Rollback/Not removed)
	// Note: Since Enforcer operations might not be fully transactional with DB if using different connection/adapter logic,
	// checking this verifies if our "Clean up Casbin" step was also rolled back or never committed (if supported)
	// OR if it was executed but "Remove" failure didn't happen because logic failed at Audit step?
	// Wait, if Audit is the LAST step, then User Delete and Casbin Remove happened.
	// If User Delete is DB Tx, it rolls back.
	// If Casbin Remove is DB Tx (same connection), it rolls back.
	// If Casbin Remove is NOT transactional, it might persist!
	// This is the CRITICAL check for "Cross-Module" integrity.
	
	rolesAfter, err := env.Enforcer.GetRolesForUser(user.ID)
	assert.NoError(t, err)
	// If Casbin adapter supports Tx, this should still contain "role:user".
	// If not, we might have an issue where User exists but Role is gone.
	assert.Contains(t, rolesAfter, "role:user", "Roles should be restored/preserved on rollback")
}
