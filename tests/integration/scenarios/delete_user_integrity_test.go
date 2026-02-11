//go:build integration
// +build integration

package scenarios

import (
	"context"
	"errors"
	"testing"
	"time"

	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	auditUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
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

func TestScenario_TransactionalIntegrity_DeleteRollback(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)

	realAuditRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)
	realAuditUC := auditUC.NewAuditUseCase(realAuditRepo, env.Logger, nil)

	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 60, 60)
	oRepo := orgRepo.NewOrganizationRepository(env.DB)
	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, oRepo, tm, env.Logger, nil, nil, env.Enforcer, realAuditUC, nil)

	setupService := userUC.NewUserUseCase(tm, env.Logger, uRepo, env.Enforcer, realAuditUC, authService, nil)
	regReq := &userModel.RegisterUserRequest{
		Username: "todelete", Email: "delete@test.com", Password: "Pass123!", Name: "To Delete",
	}
	userResp, err := setupService.Create(context.Background(), regReq)
	require.NoError(t, err)

	user, err := uRepo.FindByID(context.Background(), userResp.ID)
	require.NoError(t, err)
	require.NotNil(t, user)

	roles, err := env.Enforcer.GetRolesForUser(user.ID, "global")
	require.NoError(t, err)
	require.Contains(t, roles, "role:user")

	mockAuditUC := new(mocks.MockAuditUseCase)
	mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("intentional audit failure"))

	targetService := userUC.NewUserUseCase(tm, env.Logger, uRepo, env.Enforcer, mockAuditUC, authService, nil)

	delReq := &userModel.DeleteUserRequest{ID: user.ID}
	err = targetService.DeleteUser(context.Background(), "admin-id", delReq)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal server error")

	userAfter, err := uRepo.FindByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, userAfter)
	assert.Equal(t, user.ID, userAfter.ID)

	rolesAfter, err := env.Enforcer.GetRolesForUser(user.ID, "global")
	assert.NoError(t, err)

	assert.Contains(t, rolesAfter, "role:user", "Roles should be restored/preserved on rollback")
}
