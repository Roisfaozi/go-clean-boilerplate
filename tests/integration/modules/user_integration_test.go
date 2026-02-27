package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserIntegration(env *setup.TestEnvironment) usecase.UserUseCase {
	repo := repository.NewUserRepository(env.DB, env.Logger)

	// Use mock Audit/Auth/Storage for User UC integration to focus on User logic + DB
	auditUC := new(auditMocks.MockAuditUseCase)
	authUC := new(authMocks.MockAuthUseCase) // This is from auth/test/mocks
	storage := new(storageMocks.MockProvider)

	// However, to test Casbin integration, we pass real Enforcer
	return usecase.NewUserUseCase(
		env.TM,
		env.Logger,
		repo,
		env.Enforcer,
		auditUC,
		authUC,
		storage,
	)
}

// Helper for Auth mocks since we need them to satisfy interfaces but don't want real logic for User tests
type mockAuthUC struct { authUsecase.AuthUseCase } // Embed interface
// Implement methods needed if any called by UserUC (e.g. RevokeAllSessions)
// For integration test, we might mock this manually or use testify mock if we pass it in.
// In setupUserIntegration above we used testify mock.

func TestUserIntegration_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	uc := setupUserIntegration(env)
	ctx := context.Background()

	unique := fmt.Sprintf("crud_%d", time.Now().UnixNano())

	// 1. Create
	createReq := &model.RegisterUserRequest{
		Username: unique,
		Email:    fmt.Sprintf("%s@test.com", unique),
		Password: "Password123!",
		Name:     "Integration User",
	}

	// We need to setup expectation on the mock AuditUC if it's called
	// UserUC.Create calls AuditUC.LogActivity.
	// Since we passed a MockAuditUseCase, we need to configure it.
	// But `env` doesn't expose the mock we created inside `setupUserIntegration`.
	// Refactoring setup to return mocks or use a more integration-friendly approach.

	// RE-SETUP for this test to access mocks
	repo := repository.NewUserRepository(env.DB, env.Logger)
	auditUC := new(auditMocks.MockAuditUseCase)
	authUC := new(authMocks.MockAuthUseCase) // Note: ensure this import path is correct in file imports
	storage := new(storageMocks.MockProvider)

	// Allow any Audit call
	auditUC.On("LogActivity", ctx, mock.Anything).Return(nil)

	userUC := usecase.NewUserUseCase(env.TM, env.Logger, repo, env.Enforcer, auditUC, authUC, storage)

	user, err := userUC.Create(ctx, createReq)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, unique, user.Username)

	// Verify Casbin Role Assigned
	roles, err := env.Enforcer.GetRolesForUser(user.ID, "global")
	require.NoError(t, err)
	assert.Contains(t, roles, "role:user")

	// 2. Read
	fetched, err := userUC.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Name, fetched.Name)

	// 3. Update
	updateReq := &model.UpdateUserRequest{
		ID:   user.ID,
		Name: "Updated Name",
	}
	updated, err := userUC.Update(ctx, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)

	// 4. Delete
	// Mock backup roles logic in DeleteUser (it calls GetRolesForUser on Enforcer - real)
	// Mock RemoveFilteredGroupingPolicy - real
	// Mock Audit Log - mocked
	// Mock AddGroupingPolicy (rollback) - real

	deleteReq := &model.DeleteUserRequest{ID: user.ID}
	err = userUC.DeleteUser(ctx, "admin", deleteReq)
	require.NoError(t, err)

	// Verify Deletion (Soft Delete check)
	_, err = userUC.GetUserByID(ctx, user.ID)
	assert.Error(t, err) // Should be Not Found

	// Verify Casbin Role Removed
	roles, err = env.Enforcer.GetRolesForUser(user.ID, "global")
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestUserIntegration_DynamicSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	env := setup.SetupIntegrationEnvironment(t)
	// Minimal setup (no mocks needed for search usually)
	repo := repository.NewUserRepository(env.DB, env.Logger)
	uc := usecase.NewUserUseCase(env.TM, env.Logger, repo, nil, nil, nil, nil)
	ctx := context.Background()

	// Seed users
	prefix := fmt.Sprintf("search_%d", time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		u := &entity.User{
			ID:       fmt.Sprintf("%s_%d", prefix, i),
			Username: fmt.Sprintf("%s_user_%d", prefix, i),
			Email:    fmt.Sprintf("%s_user_%d@test.com", prefix, i),
			Name:     fmt.Sprintf("Search User %d", i),
			Status:   entity.UserStatusActive,
		}
		env.DB.Create(u)
	}

	// Filter
	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"username": {Type: "contains", From: prefix},
		},
		Sort: []querybuilder.Sort{
			{Field: "username", Desc: false},
		},
		Pagination: querybuilder.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	results, total, err := uc.GetAllUsersDynamic(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, results, 3)
}
