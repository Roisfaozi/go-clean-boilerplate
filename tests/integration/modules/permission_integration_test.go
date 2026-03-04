//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPermissionIntegration(env *setup.TestEnvironment) usecase.IPermissionUseCase {
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)
	return usecase.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)
}

func TestPermissionIntegration_AssignRoleToUser(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	user := setup.CreateTestUser(t, env.DB, "testuser_perm", "test@perm.com", "Password123!")
	roleName := "admin"
	setup.CreateTestRole(t, env.DB, roleName)

	err := permUC.AssignRoleToUser(context.Background(), user.ID, roleName, "global")
	assert.NoError(t, err)

	roles, err := env.Enforcer.GetRolesForUser(user.ID, "global")
	assert.NoError(t, err)
	assert.Contains(t, roles, roleName)
}

func TestPermissionIntegration_GrantPermission(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	roleName := "editor"
	setup.CreateTestRole(t, env.DB, roleName)

	err := permUC.GrantPermissionToRole(context.Background(), roleName, "/api/v1/articles", "POST", "global")
	assert.NoError(t, err)

	ok, _ := env.Enforcer.Enforce(roleName, "global", "/api/v1/articles", "POST")
	assert.True(t, ok)
}

func TestPermissionIntegration_RevokePermission(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	roleName := "viewer"
	setup.CreateTestRole(t, env.DB, roleName)
	_, _ = env.Enforcer.AddPolicy(roleName, "global", "/api/v1/articles", "GET")

	err := permUC.RevokePermissionFromRole(context.Background(), roleName, "/api/v1/articles", "GET", "global")
	assert.NoError(t, err)

	ok, _ := env.Enforcer.Enforce(roleName, "global", "/api/v1/articles", "GET")
	assert.False(t, ok)
}

func TestPermissionIntegration_UpdatePermission(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	roleName := "manager"
	setup.CreateTestRole(t, env.DB, roleName)
	oldP := []string{roleName, "global", "/api/v1/old", "GET"}
	newP := []string{roleName, "global", "/api/v1/new", "POST"}

	_, _ = env.Enforcer.AddPolicy(oldP[0], oldP[1], oldP[2], oldP[3])

	ok, err := permUC.UpdatePermission(context.Background(), oldP, newP)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, _ = env.Enforcer.Enforce(newP[0], "global", newP[2], newP[3])
	assert.True(t, ok)
}

func TestPermissionIntegration_GetAllPermissions(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	_, err := permUC.GetAllPermissions(context.Background())
	assert.NoError(t, err)
}

func TestPermissionIntegration_GetPermissionsForRole(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	roleName := "role_for_list"
	setup.CreateTestRole(t, env.DB, roleName)
	_, _ = env.Enforcer.AddPolicy(roleName, "global", "/res", "GET")

	policies, err := permUC.GetPermissionsForRole(context.Background(), roleName)
	assert.NoError(t, err)
	assert.NotEmpty(t, policies)
}

func TestPermissionIntegration_FullLifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	roleName := "lifecycle_role"
	setup.CreateTestRole(t, env.DB, roleName)

	err := permUC.GrantPermissionToRole(context.Background(), roleName, "/api/v1/data", "GET", "global")
	require.NoError(t, err)

	oldP := []string{roleName, "global", "/api/v1/data", "GET"}
	newP := []string{roleName, "global", "/api/v1/data/updated", "POST"}
	_, err = permUC.UpdatePermission(context.Background(), oldP, newP)
	require.NoError(t, err)

	err = permUC.RevokePermissionFromRole(context.Background(), roleName, "/api/v1/data/updated", "POST", "global")
	require.NoError(t, err)

	policies, _ := permUC.GetPermissionsForRole(context.Background(), roleName)
	assert.Empty(t, policies)
}

func TestPermissionIntegration_Negative_GrantNonExistentRole(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	err := permUC.GrantPermissionToRole(context.Background(), "non_existent_role", "/any", "GET", "global")
	assert.Error(t, err)
}

func TestPermissionIntegration_Negative_AssignRoleToNonExistentUser(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	setup.CreateTestRole(t, env.DB, "valid_role")

	err := permUC.AssignRoleToUser(context.Background(), "non-existent-user-id", "valid_role", "global")

	assert.Error(t, err)
}

func TestPermissionIntegration_Negative_RevokeNonExistentPermission(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	permUC := setupPermissionIntegration(env)
	roleName := "valid_role"
	setup.CreateTestRole(t, env.DB, roleName)

	err := permUC.RevokePermissionFromRole(context.Background(), roleName, "/ghost", "GET", "global")

	_ = err
}
