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

// ============================================
// POSITIVE TEST CASES
// ============================================

func TestPermissionIntegration_GrantPermission_Positive_ValidGrant(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	permUC := setupPermissionUseCase(t, env)
	err := permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "GET")
	require.NoError(t, err)

	policies, err := env.Enforcer.GetFilteredPolicy(0, "role:admin", "/api/v1/users", "GET")
	require.NoError(t, err)
	assert.NotEmpty(t, policies)
}

func TestPermissionIntegration_AssignRole_Positive_ValidAssignment(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")
	permUC := setupPermissionUseCase(t, env)

	err := permUC.AssignRoleToUser(context.Background(), testUser.ID, "role:admin")
	require.NoError(t, err)

	roles, err := env.Enforcer.GetRolesForUser(testUser.ID)
	require.NoError(t, err)
	assert.Contains(t, roles, "role:admin")
}

func TestPermissionIntegration_GetAllPermissions_Positive_MultiplePermissions(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	permUC := setupPermissionUseCase(t, env)
	_ = permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "GET")
	_ = permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "POST")

	result, err := permUC.GetAllPermissions()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 2)
}

// ============================================
// NEGATIVE TEST CASES
// ============================================

func TestPermissionIntegration_AssignRole_Negative_NonExistentRole(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	testUser := setup.CreateTestUser(t, env.DB, "testuser", "test@example.com", "password123")
	permUC := setupPermissionUseCase(t, env)
	err := permUC.AssignRoleToUser(context.Background(), testUser.ID, "role:nonexistent")
	assert.Error(t, err)
}

func TestPermissionIntegration_RevokePermission_Negative_NonExistentPermission(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	permUC := setupPermissionUseCase(t, env)
	err := permUC.RevokePermissionFromRole(context.Background(), "role:admin", "/api/v1/nonexistent", "GET")
	assert.Error(t, err)
}

func TestPermissionIntegration_GrantPermission_Negative_EmptyRole(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	permUC := setupPermissionUseCase(t, env)
	err := permUC.GrantPermissionToRole(context.Background(), "", "/api/v1/users", "GET")
	assert.Error(t, err)
}

func TestPermissionIntegration_GrantPermission_Negative_EmptyPath(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	permUC := setupPermissionUseCase(t, env)
	err := permUC.GrantPermissionToRole(context.Background(), "role:admin", "", "GET")
	assert.Error(t, err)
}

func TestPermissionIntegration_GrantPermission_Negative_EmptyMethod(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	permUC := setupPermissionUseCase(t, env)
	err := permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "")
	assert.Error(t, err)
}

// ============================================
// SECURITY TEST CASES
// ============================================

func TestPermissionIntegration_Security_PrivilegeEscalation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:user" })
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	regularUser := setup.CreateTestUser(t, env.DB, "regular", "regular@example.com", "password123")
	permUC := setupPermissionUseCase(t, env)

	err := permUC.AssignRoleToUser(context.Background(), regularUser.ID, "role:user")
	assert.NoError(t, err)

	err = permUC.AssignRoleToUser(context.Background(), regularUser.ID, "role:admin")
	assert.NoError(t, err)
}

func TestPermissionIntegration_Security_DuplicatePermissionGrant(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	permUC := setupPermissionUseCase(t, env)
	_ = permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "GET")
	err := permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "GET")
	assert.NoError(t, err) // Casbin normally handles duplicates by doing nothing
}

func TestPermissionIntegration_Security_CaseSensitiveRoles(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	roleFactory := fixtures.NewRoleFactory(env.DB)
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:Admin" })
	roleFactory.Create(func(r *roleEntity.Role) { r.Name = "role:admin" })

	permUC := setupPermissionUseCase(t, env)
	err1 := permUC.GrantPermissionToRole(context.Background(), "role:Admin", "/api/v1/users", "GET")
	err2 := permUC.GrantPermissionToRole(context.Background(), "role:admin", "/api/v1/users", "GET")
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func setupPermissionUseCase(t *testing.T, env *setup.TestEnvironment) usecase.IPermissionUseCase {
	roleRepo := roleRepository.NewRoleRepository(env.DB, env.Logger)
	return usecase.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo)
}