package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPermissionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env := setup.SetupIntegrationEnvironment(t)
	setup.CleanupDatabase(t, env.DB)

	rRepo := roleRepo.NewRoleRepository(env.DB, logrus.New())
	permUC := usecase.NewPermissionUseCase(env.Enforcer, logrus.New(), rRepo)

	t.Run("Assign Role to User", func(t *testing.T) {
		user := setup.CreateTestUser(t, env.DB, "testuser_perm", "test@perm.com", "Password123!")
		roleName := "admin"
		setup.CreateTestRole(t, env.DB, roleName)

		err := permUC.AssignRoleToUser(context.Background(), user.ID, roleName)
		assert.NoError(t, err)

		roles, err := env.Enforcer.GetRolesForUser(user.ID)
		assert.NoError(t, err)
		assert.Contains(t, roles, roleName)
	})

	t.Run("Grant Permission to Role", func(t *testing.T) {
		roleName := "editor"
		setup.CreateTestRole(t, env.DB, roleName)

		err := permUC.GrantPermissionToRole(context.Background(), roleName, "/api/v1/articles", "POST")
		assert.NoError(t, err)

		ok, err := env.Enforcer.Enforce(roleName, "/api/v1/articles", "POST")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Revoke Permission from Role", func(t *testing.T) {
		roleName := "viewer"
		setup.CreateTestRole(t, env.DB, roleName)
		_, _ = env.Enforcer.AddPolicy(roleName, "/api/v1/articles", "GET")

		err := permUC.RevokePermissionFromRole(context.Background(), roleName, "/api/v1/articles", "GET")
		assert.NoError(t, err)

		ok, err := env.Enforcer.Enforce(roleName, "/api/v1/articles", "GET")
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("Update Permission", func(t *testing.T) {
		roleName := "manager"
		setup.CreateTestRole(t, env.DB, roleName)
		oldP := []string{roleName, "/api/v1/old", "GET"}
		newP := []string{roleName, "/api/v1/new", "POST"}

		_, _ = env.Enforcer.AddPolicy(oldP[0], oldP[1], oldP[2])

		ok, err := permUC.UpdatePermission(context.Background(), oldP, newP)
		assert.NoError(t, err)
		assert.True(t, ok)

		// Verify old is gone
		ok, _ = env.Enforcer.Enforce(oldP[0], oldP[1], oldP[2])
		assert.False(t, ok)

		// Verify new is present
		ok, _ = env.Enforcer.Enforce(newP[0], newP[1], newP[2])
		assert.True(t, ok)
	})

	t.Run("GetAllPermissions", func(t *testing.T) {
		_, err := permUC.GetAllPermissions(context.Background())
		assert.NoError(t, err)
	})

	t.Run("GetPermissionsForRole", func(t *testing.T) {
		roleName := "role_for_list"
		setup.CreateTestRole(t, env.DB, roleName)
		_, _ = env.Enforcer.AddPolicy(roleName, "/res", "GET")

		policies, err := permUC.GetPermissionsForRole(context.Background(), roleName)
		assert.NoError(t, err)
		assert.NotEmpty(t, policies)
	})
}
