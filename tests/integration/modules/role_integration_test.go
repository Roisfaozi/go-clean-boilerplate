//go:build integration_legacy
// +build integration_legacy

package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleIntegration_Create_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	req := &model.CreateRoleRequest{
		ID:          "role:test",
		Name:        "Test Role",
		Description: "Test role description",
	}

	result, err := roleUC.Create(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.ID, result.ID)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Description, result.Description)

	// Verify in database
	role, err := roleRepo.FindByID(context.Background(), result.ID)
	require.NoError(t, err)
	assert.Equal(t, req.Name, role.Name)
}

func TestRoleIntegration_Create_DuplicateID(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	// Create first role
	req1 := &model.CreateRoleRequest{
		ID:          "role:duplicate",
		Name:        "First Role",
		Description: "First role description",
	}

	_, err := roleUC.Create(context.Background(), req1)
	require.NoError(t, err)

	// Try to create duplicate
	req2 := &model.CreateRoleRequest{
		ID:          "role:duplicate",
		Name:        "Second Role",
		Description: "Second role description",
	}

	result, err := roleUC.Create(context.Background(), req2)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRoleIntegration_Update_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	// Create role first
	createReq := &model.CreateRoleRequest{
		ID:          "role:update",
		Name:        "Original Name",
		Description: "Original description",
	}

	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	// Update role
	updateReq := &model.UpdateRoleRequest{
		ID:          created.ID,
		Name:        "Updated Name",
		Description: "Updated description",
	}

	result, err := roleUC.Update(context.Background(), updateReq)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Name", result.Name)
	assert.Equal(t, "Updated description", result.Description)

	// Verify in database
	role, err := roleRepo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", role.Name)
}

func TestRoleIntegration_Delete_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	// Create role first
	createReq := &model.CreateRoleRequest{
		ID:          "role:delete",
		Name:        "Delete Role",
		Description: "Role to be deleted",
	}

	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	// Delete role
	err = roleUC.Delete(context.Background(), created.ID)
	require.NoError(t, err)

	// Verify deleted
	_, err = roleRepo.FindByID(context.Background(), created.ID)
	assert.Error(t, err)
}

func TestRoleIntegration_GetByID_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	// Create role first
	createReq := &model.CreateRoleRequest{
		ID:          "role:get",
		Name:        "Get Role",
		Description: "Role to retrieve",
	}

	created, err := roleUC.Create(context.Background(), createReq)
	require.NoError(t, err)

	// Get role by ID
	result, err := roleUC.GetByID(context.Background(), created.ID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, created.Name, result.Name)
}

func TestRoleIntegration_GetAll_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	// Create multiple roles
	roles := []model.CreateRoleRequest{
		{ID: "role:test1", Name: "Test Role 1", Description: "Description 1"},
		{ID: "role:test2", Name: "Test Role 2", Description: "Description 2"},
		{ID: "role:test3", Name: "Test Role 3", Description: "Description 3"},
	}

	for _, req := range roles {
		_, err := roleUC.Create(context.Background(), &req)
		require.NoError(t, err)
	}

	// Get all roles
	result, err := roleUC.GetAll(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	// Should have at least 3 test roles + 3 default roles (admin, user, moderator)
	assert.GreaterOrEqual(t, len(result), 6)
}

func TestRoleIntegration_DynamicSearch_Success(t *testing.T) {

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	roleUC := usecase.NewRoleUseCase(roleRepo, tm, env.Logger)

	// Create test roles
	roles := []model.CreateRoleRequest{
		{ID: "role:manager", Name: "Manager", Description: "Manager role"},
		{ID: "role:developer", Name: "Developer", Description: "Developer role"},
		{ID: "role:designer", Name: "Designer", Description: "Designer role"},
	}

	for _, req := range roles {
		_, err := roleUC.Create(context.Background(), &req)
		require.NoError(t, err)
	}

	// Search with filter
	filter := &model.DynamicSearchRequest{
		Filter: map[string]interface{}{
			"name": map[string]interface{}{
				"type": "contains",
				"from": "Dev",
			},
		},
	}

	result, err := roleUC.DynamicSearch(context.Background(), filter)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Developer", result[0].Name)
}
