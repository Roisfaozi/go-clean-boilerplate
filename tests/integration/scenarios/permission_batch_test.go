//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	permissionModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	permissionUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	roleUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_PermissionBatchCheck(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	ctx := context.Background()
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	// 1. Setup Services
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo)

	// 2. Create Role & User
	roleName := "Editor"
	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: roleName})
	require.NoError(t, err)

	user := setup.CreateTestUser(t, env.DB, "editor_user", "editor@batch.com", "pass")
	err = permService.AssignRoleToUser(ctx, user.ID, roleName)
	require.NoError(t, err)

	// 3. Grant Permissions
	// Editor can READ and WRITE articles, but NOT DELETE
	err = permService.GrantPermissionToRole(ctx, roleName, "/articles", "READ")
	require.NoError(t, err)
	err = permService.GrantPermissionToRole(ctx, roleName, "/articles", "WRITE")
	require.NoError(t, err)

	// 4. Perform Batch Check
	items := []permissionModel.PermissionCheckItem{
		{Resource: "/articles", Action: "READ"},
		{Resource: "/articles", Action: "WRITE"},
		{Resource: "/articles", Action: "DELETE"},
		{Resource: "/users", Action: "READ"}, // Unrelated resource
	}

	results, err := permService.BatchCheckPermission(ctx, user.ID, items)
	require.NoError(t, err)

	// 5. Verify Results
	assert.True(t, results["/articles:READ"], "Should be able to READ articles")
	assert.True(t, results["/articles:WRITE"], "Should be able to WRITE articles")
	assert.False(t, results["/articles:DELETE"], "Should NOT be able to DELETE articles")
	assert.False(t, results["/users:READ"], "Should NOT be able to READ users")
}
