//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

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

// TestScenario_RoleHierarchy verifies that role inheritance allows
// a parent role to access resources granted to a child role.
func TestScenario_RoleHierarchy(t *testing.T) {
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

	// 2. Create Roles
	parentRole := "Manager"
	childRole := "Staff"

	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: parentRole})
	require.NoError(t, err)
	_, err = roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: childRole})
	require.NoError(t, err)

	// 3. Grant Permission to Child Role ONLY
	path := "/api/v1/work"
	method := "GET"
	err = permService.GrantPermissionToRole(ctx, childRole, path, method)
	require.NoError(t, err)

	// 4. Verify Parent cannot access yet
	ok, err := env.Enforcer.Enforce(parentRole, path, method)
	require.NoError(t, err)
	assert.False(t, ok, "Parent role should not have access yet")

	// 5. Add Inheritance: Parent inherits from Child
	// g(parent, child) -> Parent is a member of Child group?
	// Wait, Casbin RBAC 'g' is: g(user, role).
	// For role hierarchy: g(role1, role2) means role1 is a member of role2.
	// So if role1 inherits role2, role1 should have role2's permissions.
	// Example: alice is admin. admin is user. alice can do what user can do.
	// g(alice, admin)
	// g(admin, user)
	// p(user, data, read)
	// enforce(alice, data, read) -> true.
	
	// So: AddParentRole(childRole, parentRole) in my implementation calls AddGroupingPolicy(child, parent).
	// Wait, in my implementation: AddGroupingPolicy(childRole, parentRole).
	// If I call AddParentRole("Staff", "Manager"), it executes g("Staff", "Manager").
	// This means "Staff" is a member of "Manager".
	// Usually hierarchy means "Manager" HAS "Staff" permissions.
	// So "Manager" should be the "subject" that is member of "Staff" group?
	// Let's re-read Casbin logic.
	
	// If p(Staff, data, read).
	// We want Manager to read data.
	// So Manager must be "part of" Staff.
	// g(Manager, Staff).
	// This means Manager is a "member" of Staff role.
	
	// My implementation: AddParentRole(childRole, parentRole) -> g(childRole, parentRole).
	// If I want Manager (Parent) to inherit Staff (Child), I should call:
	// AddParentRole("Manager", "Staff") -> g("Manager", "Staff").
	
	// But usually "Parent" implies "Higher Level".
	// "Manager" is Parent of "Staff".
	// So "Manager" > "Staff".
	// Inheritance direction: Higher Role inherits Lower Role.
	
	// Let's verify the naming in implementation:
	// func (uc *PermissionUseCase) AddParentRole(ctx context.Context, childRole, parentRole string) error {
	//    uc.enforcer.AddGroupingPolicy(childRole, parentRole)
	// }
	
	// If I call AddParentRole("Manager", "Staff"), it creates g("Manager", "Staff").
	// Then Manager has Staff permissions.
	// Is "Manager" the child or parent in variable naming?
	// In the function signature: `childRole` is the first arg.
	// So `childRole` is the subject. `parentRole` is the group.
	// If I want Manager to have Staff rights, Manager is the subject. Staff is the group.
	// So `childRole` = Manager (The one inheriting), `parentRole` = Staff (The one being inherited).
	
	// Wait, "Child inherits from Parent" usually means Child gets Parent's traits.
	// But in RBAC hierarchy, "Admin" (Super) inherits "User" (Base).
	// Admin is "above" User.
	// So Admin is the "Child" (in OO terms) of "User"? No, that's confusing.
	
	// Let's stick to Casbin `g(u, r)` semantic: `u` is member of `r`.
	// If we want `u` to have `r` permissions.
	// If we want `RoleA` to have `RoleB` permissions.
	// We need `g(RoleA, RoleB)`.
	// `RoleA` is the subject. `RoleB` is the group/role.
	
	// My function: `AddParentRole(childRole, parentRole)` -> `g(childRole, parentRole)`.
	// So `childRole` becomes member of `parentRole`.
	// `childRole` gets `parentRole` permissions.
	
	// In this test:
	// ParentRole = "Manager". ChildRole = "Staff".
	// We want Manager to have Staff permissions.
	// So we need `g(Manager, Staff)`.
	// So we call `AddParentRole("Manager", "Staff")`.
	// Here "Manager" is the `childRole` (inheritor) and "Staff" is the `parentRole` (source).
	// The naming `childRole` / `parentRole` in my implementation might be slightly misleading if mapped to OO.
	// In OO: Child extends Parent. Child has Parent features.
	// So `Manager` (Child) extends `Staff` (Parent).
	// So `AddParentRole(Manager, Staff)` makes sense.
	
	err = permService.AddParentRole(ctx, parentRole, childRole) 
	require.NoError(t, err)

	// 6. Verify Parent HAS access now
	// We might need to LoadPolicy or wait if using a watcher (but here it's direct DB adapter usually)
	// However, Enforcer in memory should be updated.
	
	ok, err = env.Enforcer.Enforce(parentRole, path, method)
	require.NoError(t, err)
	assert.True(t, ok, "Parent role (Manager) should inherit permissions from Child role (Staff)")
}
