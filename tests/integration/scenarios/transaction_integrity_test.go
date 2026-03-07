//go:build integration
// +build integration

package scenarios

import (
	"context"
	"errors"
	"testing"
	"time"

	auditRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	userUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sso"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScenario_TransactionalIntegrity_RegisterRollback(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	mockEnforcer := new(mocks.MockIEnforcer)

	// Mock WithContext to return itself
	mockEnforcer.On("WithContext", mock.Anything).Return(mockEnforcer)

	tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	aucRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)
	auditService := auditUC.NewAuditUseCase(aucRepo, env.Logger, nil, nil)
	jwtManager := jwt.NewJWTManager("secret", "refresh", 60, 60)
	oRepo := orgRepo.NewOrganizationRepository(env.DB)
	authz := authRepo.NewCasbinAdapter(mockEnforcer, "role:user", "global")
	authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, oRepo, tm, env.Logger, nil, authz, nil, nil, make(map[string]sso.Provider))

	userService := userUC.NewUserUseCase(tm, env.Logger, uRepo, mockEnforcer, auditService, authService, nil)

	expectedErr := errors.New("casbin connection error")

	// My manual mock uses variadic params...interface{} which mockery packs into a slice.
	mockEnforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
		if len(params) != 3 {
			return false
		}
		// Check if the last param is "global" as expected in the test
		return params[2] == "global"
	})).Return(false, expectedErr)

	req := &userModel.RegisterUserRequest{
		Username: "rollback_user",
		Email:    "rollback@test.com",
		Password: "Password123!",
		Name:     "Rollback User",
	}

	_, err := userService.Create(context.Background(), req)

	require.Error(t, err, "Expected error from UserUseCase when Role assignment fails")

	user, _ := uRepo.FindByUsername(context.Background(), req.Username)
	assert.Nil(t, user, "User should be rolled back (not found) when role assignment fails")
}
