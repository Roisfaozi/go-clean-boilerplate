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

func TestPermissionIntegration_Comprehensive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	env := setup.SetupIntegrationEnvironment(t)
	setup.CleanupDatabase(t, env.DB)

	rRepo := roleRepo.NewRoleRepository(env.DB, logrus.New())
	permUC := usecase.NewPermissionUseCase(env.Enforcer, logrus.New(), rRepo)

	t.Run("Full Permission Lifecycle", func(t *testing.T) {
		roleName := "comprehensive_role"
		setup.CreateTestRole(t, env.DB, roleName)

		// 1. Grant
		err := permUC.GrantPermissionToRole(context.Background(), roleName, "/api/v1/data", "GET")
		assert.NoError(t, err)

		// 2. Verify Grant
		policies, err := permUC.GetPermissionsForRole(context.Background(), roleName)
		assert.NoError(t, err)
		assert.Len(t, policies, 1)
		assert.Equal(t, []string{roleName, "/api/v1/data", "GET"}, policies[0])

		// 3. Update
		oldP := []string{roleName, "/api/v1/data", "GET"}
		newP := []string{roleName, "/api/v1/data/updated", "POST"}
		ok, err := permUC.UpdatePermission(context.Background(), oldP, newP)
		assert.NoError(t, err)
		assert.True(t, ok)

		// 4. Verify Update
		policies, err = permUC.GetPermissionsForRole(context.Background(), roleName)
		assert.NoError(t, err)
		assert.Len(t, policies, 1)
		assert.Equal(t, newP, policies[0])

		// 5. Revoke
		err = permUC.RevokePermissionFromRole(context.Background(), roleName, "/api/v1/data/updated", "POST")
		assert.NoError(t, err)

		// 6. Final Verify
		policies, err = permUC.GetPermissionsForRole(context.Background(), roleName)
		assert.NoError(t, err)
		assert.Empty(t, policies)
	})

	t.Run("Bulk Operations and Security", func(t *testing.T) {
		// Test GetAllPermissions
		_, err := permUC.GetAllPermissions(context.Background())
		assert.NoError(t, err)

		// Test invalid inputs
		err = permUC.GrantPermissionToRole(context.Background(), "non_existent_role", "/any", "GET")
		assert.Error(t, err, "Should fail for non-existent role")
	})
}