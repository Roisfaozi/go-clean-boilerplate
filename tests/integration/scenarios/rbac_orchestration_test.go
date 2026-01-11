//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	accessRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	accessUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
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

// Scenario: RBAC Orchestration (Full Chain)
// Verifies: Role -> Endpoint -> AccessRight -> Link -> Grant -> Assign -> Enforce
func TestScenario_RBAC_Orchestration(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	ctx := context.Background()
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	// 1. Setup Modules
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo)

	aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)
	accessService := accessUC.NewAccessUseCase(aRepo, env.Logger)

	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo)

	// 2. Create Role "Analyst"
	roleName := "Analyst"
	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: roleName, Description: "Data Analyst"})
	require.NoError(t, err)

	// 3. Create Endpoint "GET /api/v1/reports"
	endpoint, err := accessService.CreateEndpoint(ctx, accessModel.CreateEndpointRequest{
		Path:   "/api/v1/reports",
		Method: "GET",
	})
	require.NoError(t, err)

	// 4. Create Access Right "View Reports"
	accessRight, err := accessService.CreateAccessRight(ctx, accessModel.CreateAccessRightRequest{
		Name:        "view_reports",
		Description: "Can view daily reports",
	})
	require.NoError(t, err)

	// 5. Link Endpoint to Access Right
	err = accessService.LinkEndpointToAccessRight(ctx, accessModel.LinkEndpointRequest{
		AccessRightID: accessRight.ID,
		EndpointID:    endpoint.ID,
	})
	require.NoError(t, err)

	// 6. Grant Permission (Orchestration Step)
	// In a real app, the controller would look up all endpoints for "view_reports" and grant them.
	// Here we simulate that logic by iterating the endpoints of the access right.
	// Note: AccessUseCase currently doesn't expose "GetEndpointsForAccessRight", 
	// so we will query the endpoint directly as we know it.
	
	// Validating that we can grant the specific path/method defined in the endpoint
	err = permService.GrantPermissionToRole(ctx, roleName, endpoint.Path, endpoint.Method)
	require.NoError(t, err)

	// 7. Create User and Assign Role
	user := setup.CreateTestUser(t, env.DB, "analyst_user", "analyst@test.com", "pass")
	err = permService.AssignRoleToUser(ctx, user.ID, roleName)
	require.NoError(t, err)

	// 8. Verify Access via Enforcer
	// User should have access to the endpoint path/method because they have the role
	ok, err := env.Enforcer.Enforce(user.ID, endpoint.Path, endpoint.Method)
	require.NoError(t, err)
	assert.True(t, ok, "User should be able to access the endpoint granted via role")

	// 9. Verify Denial
	// User should NOT have access to DELETE
	ok, _ = env.Enforcer.Enforce(user.ID, endpoint.Path, "DELETE")
	assert.False(t, ok, "User should not have DELETE permission")
}
