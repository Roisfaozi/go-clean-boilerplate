package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	accessMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type permissionTestDeps struct {
	Enforcer   *mocks.MockIEnforcer
	RoleRepo   *roleMocks.MockRoleRepository
	UserRepo   *userMocks.MockUserRepository
	AccessRepo *accessMocks.MockAccessRepository
	AuditUC    *auditMocks.MockAuditUseCase
}

func setupPermissionTest() (*permissionTestDeps, usecase.IPermissionUseCase) {
	deps := &permissionTestDeps{
		Enforcer:   new(mocks.MockIEnforcer),
		RoleRepo:   new(roleMocks.MockRoleRepository),
		UserRepo:   new(userMocks.MockUserRepository),
		AccessRepo: new(accessMocks.MockAccessRepository),
		AuditUC:    new(auditMocks.MockAuditUseCase),
	}

	// Default behavior for enforcer with context to return itself
	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)

	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := usecase.NewPermissionUseCase(deps.Enforcer, log, deps.RoleRepo, deps.UserRepo, deps.AccessRepo, deps.AuditUC)
	return deps, uc
}

func TestAssignRoleToUser_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID, roleName := "user123", "editor"

	// Mock UserRepo
	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)

	// Mock RoleRepo
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.NoError(t, err)
	deps.UserRepo.AssertExpectations(t)
	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestAssignRoleToUser_UserNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)

	deps.RoleRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
	deps.Enforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything)
}

func TestAssignRoleToUser_UserRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("db error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestAssignRoleToUser_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID, roleName := "user123", "non_existent_role"

	// Mock UserRepo Success
	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)

	// Mock RoleRepo Fail
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	deps.Enforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything)
}

func TestAssignRoleToUser_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(nil, errors.New("db error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestAssignRoleToUser_EnforcerRemoveError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("casbin error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestAssignRoleToUser_EnforcerAddError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName, "global")

	assert.Error(t, err)
	assert.Equal(t, errors.New("casbin error"), err)
}

func TestAssignRoleToUser_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()

	err := uc.AssignRoleToUser(context.Background(), "", "role", "global")
	assert.Error(t, err)

	err = uc.AssignRoleToUser(context.Background(), "user", "", "global")
	assert.Error(t, err)
}

func TestGrantPermissionToRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "editor", "/api/v1/articles", "POST"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method, "global")

	assert.NoError(t, err)
	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestGrantPermissionToRole_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "non_existent_role", "/api/v1/articles", "POST"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	deps.Enforcer.AssertNotCalled(t, "AddPolicy", mock.Anything, mock.Anything, mock.Anything)
}

func TestGrantPermissionToRole_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "POST"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, errors.New("db error"))

	err := uc.GrantPermissionToRole(context.Background(), role, path, method, "global")
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestGrantPermissionToRole_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "POST"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, errors.New("casbin error"))

	err := uc.GrantPermissionToRole(context.Background(), role, path, method, "global")
	assert.Error(t, err)
	assert.Equal(t, errors.New("casbin error"), err)
}

func TestGrantPermissionToRole_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	assert.Error(t, uc.GrantPermissionToRole(context.Background(), "", "path", "GET", "global"))
	assert.Error(t, uc.GrantPermissionToRole(context.Background(), "role", "", "GET", "global"))
	assert.Error(t, uc.GrantPermissionToRole(context.Background(), "role", "path", "", "global"))
}

func TestRevokePermissionFromRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "editor", "/api/v1/articles", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", mock.Anything).Return(true, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method, "global")

	assert.NoError(t, err)
	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestRevokePermissionFromRole_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "non_existent_role", "/api/v1/articles", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method, "global")

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	deps.Enforcer.AssertNotCalled(t, "RemovePolicy", mock.Anything, mock.Anything, mock.Anything)
}

func TestRevokePermissionFromRole_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, errors.New("db error"))

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method, "global")
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestRevokePermissionFromRole_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", mock.Anything).Return(false, errors.New("casbin error"))

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method, "global")
	assert.Error(t, err)
	assert.Equal(t, errors.New("casbin error"), err)
}

func TestRevokePermissionFromRole_PolicyNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", mock.Anything).Return(false, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method, "global")
	assert.Error(t, err)
	assert.Equal(t, errors.New("policy to revoke not found in specified domain"), err)
}

func TestRevokePermissionFromRole_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	assert.Error(t, uc.RevokePermissionFromRole(context.Background(), "", "path", "GET", "global"))
	assert.Error(t, uc.RevokePermissionFromRole(context.Background(), "role", "", "GET", "global"))
	assert.Error(t, uc.RevokePermissionFromRole(context.Background(), "role", "path", "", "global"))
}

func TestGetAllPermissions_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	expectedPolicies := [][]string{{"role", "path", "GET"}}

	deps.Enforcer.On("GetPolicy").Return(expectedPolicies, nil)

	policies, err := uc.GetAllPermissions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicies, policies)
}

func TestGetAllPermissions_Error(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.Enforcer.On("GetPolicy").Return(nil, errors.New("casbin error"))

	_, err := uc.GetAllPermissions(context.Background())
	assert.Error(t, err)
}

func TestGetPermissionsForRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	expectedPolicies := [][]string{{"admin", "path", "GET"}}

	deps.Enforcer.On("GetFilteredPolicy", 0, mock.Anything).Return(expectedPolicies, nil)

	policies, err := uc.GetPermissionsForRole(context.Background(), role)
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicies, policies)
}

func TestGetPermissionsForRole_Error(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	deps.Enforcer.On("GetFilteredPolicy", 0, mock.Anything).Return(nil, errors.New("casbin error"))

	_, err := uc.GetPermissionsForRole(context.Background(), role)
	assert.Error(t, err)
}

func TestUpdatePermission_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"role", "old", "GET"}
	newP := []string{"role", "new", "GET"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(true, nil)

	updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestUpdatePermission_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	updated, err := uc.UpdatePermission(context.Background(), []string{}, []string{"a"})
	assert.Error(t, err)
	assert.False(t, updated)

	updated, err = uc.UpdatePermission(context.Background(), []string{"a"}, []string{})
	assert.Error(t, err)
	assert.False(t, updated)
}

func TestUpdatePermission_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"role", "old", "GET"}
	newP := []string{"role", "new", "GET"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(false, errors.New("casbin error"))

	updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
	assert.Error(t, err)
	assert.False(t, updated)
}

func TestUpdatePermission_PolicyNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"role", "old", "GET"}
	newP := []string{"role", "new", "GET"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(false, nil)

	updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
	assert.Error(t, err)
	assert.False(t, updated)
}

// ============================================================================
// BATCH PERMISSION CHECK TESTS
// ============================================================================

func TestPermissionUseCase_BatchCheckPermission_Success_AllAllowed(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-123"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/users", Action: "POST"},
		{Resource: "/api/roles", Action: "GET"},
	}

	// Mock Enforce - All allowed
	deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.True(t, results["/api/users:POST"])
	assert.True(t, results["/api/roles:GET"])
}

func TestPermissionUseCase_BatchCheckPermission_Success_Mixed(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-456"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},    // Allowed
		{Resource: "/api/users", Action: "DELETE"}, // Denied
		{Resource: "/api/roles", Action: "POST"},   // Allowed
		{Resource: "/api/admin", Action: "GET"},    // Denied
	}

	// Mock Enforce - Mixed results
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/users" && p[3] == "GET"
	})).Return(true, nil)
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/users" && p[3] == "DELETE"
	})).Return(false, nil)
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/roles" && p[3] == "POST"
	})).Return(true, nil)
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/admin" && p[3] == "GET"
	})).Return(false, nil)
	deps.Enforcer.On("Enforce", mock.Anything).Return(false, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 4, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.False(t, results["/api/users:DELETE"])
	assert.True(t, results["/api/roles:POST"])
	assert.False(t, results["/api/admin:GET"])
}

func TestPermissionUseCase_BatchCheckPermission_EmptyItems(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-789"
	items := []model.PermissionCheckItem{} // Empty list

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Should return empty map
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
	deps.Enforcer.AssertNotCalled(t, "Enforce")
}

func TestPermissionUseCase_BatchCheckPermission_LargeItemList(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-101"

	// Create 100 items
	items := make([]model.PermissionCheckItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = model.PermissionCheckItem{
			Resource: "/api/resource",
			Action:   "GET",
		}
		// Mock each enforce call
		deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Once()
	}

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Note: All items have same resource:action, so only 1 key in map
	assert.Equal(t, 1, len(results))
	assert.True(t, results["/api/resource:GET"])
}

func TestPermissionUseCase_BatchCheckPermission_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-202"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/roles", Action: "POST"}, // This will error
		{Resource: "/api/admin", Action: "GET"},
	}

	// Mock Enforce - One with error
	deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Once()
	deps.Enforcer.On("Enforce", mock.Anything).Return(false, errors.New("casbin database error")).Once()
	deps.Enforcer.On("Enforce", mock.Anything).Return(false, nil).Once()

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Should not fail, but log error and continue
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.False(t, results["/api/roles:POST"]) // Error treated as false
	assert.False(t, results["/api/admin:GET"])
}

// ============================================================================
// GUARDIAN PERMISSION TESTS
// ============================================================================

func TestRevokeRoleFromUser_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)

	err := uc.RevokeRoleFromUser(context.Background(), userID, roleName, "global")
	assert.NoError(t, err)
}

func TestRevokeRoleFromUser_Guardian_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	assert.Error(t, uc.RevokeRoleFromUser(context.Background(), "", "role", "global"))
	assert.Error(t, uc.RevokeRoleFromUser(context.Background(), "user", "", "global"))
}

func TestRevokeRoleFromUser_Guardian_UserNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRevokeRoleFromUser_Guardian_UserRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, errors.New("db fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRevokeRoleFromUser_Guardian_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, errors.New("db fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, errors.New("casbin fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_RoleNotAssigned(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, nil) // Removed = false

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role was not assigned to user")
}

func TestAddParentRole_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.RoleRepo.On("FindByName", mock.Anything, child).Return(&roleEntity.Role{Name: child}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, parent).Return(&roleEntity.Role{Name: parent}, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AddParentRole(context.Background(), child, parent, "global")
	assert.NoError(t, err)
}

func TestAddParentRole_Guardian_ChildRoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(nil, errors.New("not found"))

	err := uc.AddParentRole(context.Background(), "child", "parent", "global")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestAddParentRole_Guardian_ParentRoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(&roleEntity.Role{Name: "child"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "parent").Return(nil, errors.New("not found"))

	err := uc.AddParentRole(context.Background(), "child", "parent", "global")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestAddParentRole_Guardian_SelfInheritance(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)

	err := uc.AddParentRole(context.Background(), "role", "role", "global")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot inherit from itself")
}

func TestAddParentRole_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.RoleRepo.On("FindByName", mock.Anything, child).Return(&roleEntity.Role{Name: child}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, parent).Return(&roleEntity.Role{Name: parent}, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin fail"))

	err := uc.AddParentRole(context.Background(), child, parent, "global")
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestRemoveParentRole_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)

	err := uc.RemoveParentRole(context.Background(), child, parent, "global")
	assert.NoError(t, err)
}

func TestRemoveParentRole_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, errors.New("casbin fail"))

	err := uc.RemoveParentRole(context.Background(), child, parent, "global")
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestRemoveParentRole_Guardian_RelationshipNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, nil)

	err := uc.RemoveParentRole(context.Background(), child, parent, "global")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inheritance relationship not found")
}

func TestGetParentRoles_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "editor"
	parents := []string{"viewer"}

	deps.Enforcer.On("GetRolesForUser", mock.Anything, mock.Anything).Return(parents, nil)

	res, err := uc.GetParentRoles(context.Background(), role, "global")
	assert.NoError(t, err)
	assert.Equal(t, parents, res)
}

func TestGetParentRoles_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "editor"

	deps.Enforcer.On("GetRolesForUser", mock.Anything, mock.Anything).Return(nil, errors.New("casbin fail"))

	_, err := uc.GetParentRoles(context.Background(), role, "global")
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestPermissionUseCase_Edge_MaxStringLength(t *testing.T) {
	deps, uc := setupPermissionTest()
	longString := strings.Repeat("a", 1000)

	deps.RoleRepo.On("FindByName", mock.Anything, longString).Return(nil, errors.New("record not found"))

	err := uc.GrantPermissionToRole(context.Background(), longString, longString, "GET", "global")
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestPermissionUseCase_Vulnerability_SQLInjectionInRole(t *testing.T) {
	deps, uc := setupPermissionTest()
	sqliRole := "admin' OR '1'='1"

	deps.RoleRepo.On("FindByName", mock.Anything, sqliRole).Return(nil, errors.New("record not found"))

	err := uc.GrantPermissionToRole(context.Background(), sqliRole, "/path", "GET", "global")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrInternalServer) || errors.Is(err, exception.ErrBadRequest))
}

func TestPermissionUseCase_Negative_AssignRoleToUser_EmptyUser(t *testing.T) {
	_, uc := setupPermissionTest()

	err := uc.AssignRoleToUser(context.Background(), "", "role", "global")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "empty"))
}

func TestPermissionUseCase_Negative_GrantPermissionToRole_SpecialChars(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	path := "/api/v1/resource/!@#$%^&*()"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), role, path, "GET", "global")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}
