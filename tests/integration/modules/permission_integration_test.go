//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionIntegration_GrantPermission_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	permUC := usecase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)

	err := permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "GET")
	require.NoError(t, err)

	policies, err := env.Enforcer.GetFilteredPolicy(0, "role:admin", "/api/v1/users", "GET")
	require.NoError(t, err)
	assert.NotEmpty(t, policies)
}

func TestPermissionIntegration_RevokePermission_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:test" })

	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	permUC := usecase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)

	err := permUC.GrantPermissionToRole(context.Background(), "role:test", "/api/v1/test", "POST")
	require.NoError(t, err)

	err = permUC.RevokePermissionFromRole(context.Background(), "role:test", "/api/v1/test", "POST")
	require.NoError(t, err)

	policies, err := env.Enforcer.GetFilteredPolicy(0, "role:test", "/api/v1/test", "POST")
	require.NoError(t, err)
	assert.Empty(t, policies)
}

func TestPermissionIntegration_AssignRole_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")

	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	permUC := usecase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)

	err := permUC.AssignRoleToUser(context.Background(), testUser.ID, "role:admin")
	require.NoError(t, err)

	roles, err := env.Enforcer.GetRolesForUser(testUser.ID)
	require.NoError(t, err)
	assert.Contains(t, roles, "role:admin")
}

func TestPermissionIntegration_GetAllPermissions_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:test" })

	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	permUC := usecase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)

	err := permUC.GrantPermissionToRole(context.Background(), "role:test", "/api/v1/users", "GET")
	require.NoError(t, err)

	permissions, err := permUC.GetAllPermissions()
	require.NoError(t, err)
	assert.NotEmpty(t, permissions)
}

func TestPermissionIntegration_GetPermissionsForRole_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:test" })

	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	permUC := usecase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)

	err := permUC.GrantPermissionToRole(context.Background(), "role:test", "/api/v1/users", "GET")
	require.NoError(t, err)
	err = permUC.GrantPermissionToRole(context.Background(), "role:test", "/api/v1/users", "POST")
	require.NoError(t, err)

	permissions, err := permUC.GetPermissionsForRole("role:test")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(permissions), 2)
}
