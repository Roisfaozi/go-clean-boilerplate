//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	permissionModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	permissionUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	roleUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
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

	rRepo := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	aRepo := accessRepo.NewAccessRepository(env.SQLXDB, env.Logger)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo, permService)

	roleName := "Editor"
	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: roleName})
	require.NoError(t, err)

	actorID := "sa-actor-id"
	_, _ = env.Enforcer.AddGroupingPolicy(actorID, "role:superadmin", "global")
	saCtx := authcontext.WithUserID(context.Background(), actorID)

	user := setup.CreateTestUser(t, env.DB, "editor_user", "editor@batch.com", "pass")
	err = permService.AssignRoleToUser(saCtx, user.ID, roleName, "global")
	require.NoError(t, err)

	err = permService.GrantPermissionToRole(saCtx, roleName, "/articles", "READ", "global")
	require.NoError(t, err)
	err = permService.GrantPermissionToRole(saCtx, roleName, "/articles", "WRITE", "global")
	require.NoError(t, err)

	items := []permissionModel.PermissionCheckItem{
		{Resource: "/articles", Action: "READ"},
		{Resource: "/articles", Action: "WRITE"},
		{Resource: "/articles", Action: "DELETE"},
		{Resource: "/users", Action: "READ"},
	}

	results, err := permService.BatchCheckPermission(ctx, user.ID, items)
	require.NoError(t, err)

	assert.True(t, results["/articles:READ"], "Should be able to READ articles")
	assert.True(t, results["/articles:WRITE"], "Should be able to WRITE articles")
	assert.False(t, results["/articles:DELETE"], "Should NOT be able to DELETE articles")
	assert.False(t, results["/users:READ"], "Should NOT be able to READ users")
}
