//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
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

func TestScenario_RoleHierarchy(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	ctx := context.Background()
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo, permService)

	parentRole := "Manager"
	childRole := "Staff"

	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: parentRole})
	require.NoError(t, err)
	_, err = roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: childRole})
	require.NoError(t, err)

	path := "/api/v1/work"
	method := "GET"
	err = permService.GrantPermissionToRole(ctx, childRole, path, method, "global")
	require.NoError(t, err)

	ok, err := env.Enforcer.Enforce(parentRole, "global", path, method)
	require.NoError(t, err)
	assert.False(t, ok, "Parent role should not have access yet")

	err = permService.AddParentRole(ctx, parentRole, childRole, "global")
	require.NoError(t, err)

	ok, err = env.Enforcer.Enforce(parentRole, "global", path, method)
	require.NoError(t, err)
	assert.True(t, ok, "Parent role (Manager) should inherit permissions from Child role (Staff)")
}
