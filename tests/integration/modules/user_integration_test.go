//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDependencies(env *setup.TestEnvironment) (usecase.UserUseCase, repository.UserRepository) {
	userRepo := repository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger)
	
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", time.Hour, time.Hour*24)
	
	authUC := authUseCase.NewAuthUsecase(jwtManager, tokenRepo, userRepo, tm, env.Logger, nil, env.Enforcer, auditUC, nil)

	return usecase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC), userRepo
}

func TestUserIntegration_Create_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	userUC, userRepo := setupTestDependencies(env)

	req := &model.RegisterUserRequest{
		Username:  "newuser",
		Email:     "newuser@example.com",
		Password:  "password123",
		Name:      "New User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Username, result.Username)
	assert.Equal(t, req.Email, result.Email)

	user, err := userRepo.FindByID(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, req.Username, user.Username)

	roles, err := env.Enforcer.GetRolesForUser(result.ID)
	require.NoError(t, err)
	assert.Contains(t, roles, "role:user")
}

func TestUserIntegration_Create_DuplicateUsername(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	setup.CreateTestUser(t, env.DB, "existinguser", "existing@example.com", "password123")

	userUC, _ := setupTestDependencies(env)

	req := &model.RegisterUserRequest{
		Username:  "existinguser",
		Email:     "newemail@example.com",
		Password:  "password123",
		Name:      "New User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUserIntegration_Update_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	userUC, userRepo := setupTestDependencies(env)

	updateReq := &model.UpdateUserRequest{
		ID:        testUser.ID,
		Name:      "Updated Name",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	result, err := userUC.Update(context.Background(), updateReq)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Name", result.Name)

	updatedUser, err := userRepo.FindByID(context.Background(), testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
}

func TestUserIntegration_Delete_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	userUC, userRepo := setupTestDependencies(env)

	deleteReq := &model.DeleteUserRequest{
		ID:        testUser.ID,
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	err := userUC.DeleteUser(context.Background(), "admin-id", deleteReq)
	require.NoError(t, err)

	_, err = userRepo.FindByID(context.Background(), testUser.ID)
	assert.Error(t, err)
}

func TestUserIntegration_GetByID_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	userUC, _ := setupTestDependencies(env)

	result, err := userUC.GetUserByID(context.Background(), testUser.ID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testUser.ID, result.ID)
	assert.Equal(t, testUser.Username, result.Username)
}