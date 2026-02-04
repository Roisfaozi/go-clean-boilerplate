package test

import (
	"context"
	"errors"
	"testing"

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
	Enforcer *mocks.IEnforcer
	RoleRepo *roleMocks.MockRoleRepository
	UserRepo *userMocks.MockUserRepository
}

func setupPermissionTest() (*permissionTestDeps, usecase.IPermissionUseCase) {
	deps := &permissionTestDeps{
		Enforcer: new(mocks.IEnforcer),
		RoleRepo: new(roleMocks.MockRoleRepository),
		UserRepo: new(userMocks.MockUserRepository),
	}

	uc := usecase.NewPermissionUseCase(deps.Enforcer, logrus.New(), deps.RoleRepo, deps.UserRepo)
	return deps, uc
}

func TestAssignRoleToUser_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID, roleName := "user123", "editor"

	// Mock UserRepo
	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)

	// Mock RoleRepo
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", "global").Return(true, nil)
	deps.Enforcer.On("AddGroupingPolicy", userID, roleName, "global").Return(true, nil)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.NoError(t, err)
	deps.UserRepo.AssertExpectations(t)
	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestAssignRoleToUser_UserNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID, roleName := "user123", "editor"

	// Mock UserRepo Fail
	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)

	deps.RoleRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
	deps.Enforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything)
}

func TestAssignRoleToUser_UserRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("db error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

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

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	deps.Enforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything)
}

func TestAssignRoleToUser_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(nil, errors.New("db error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestAssignRoleToUser_EnforcerRemoveError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", "global").Return(false, errors.New("casbin error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestAssignRoleToUser_EnforcerAddError(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", "global").Return(true, nil)
	deps.Enforcer.On("AddGroupingPolicy", userID, roleName, "global").Return(false, errors.New("casbin error"))

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, errors.New("casbin error"), err)
}

func TestAssignRoleToUser_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()

	err := uc.AssignRoleToUser(context.Background(), "", "role")
	assert.Error(t, err)

	err = uc.AssignRoleToUser(context.Background(), "user", "")
	assert.Error(t, err)
}

func TestGrantPermissionToRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "editor", "/api/v1/articles", "POST"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", role, "global", path, method).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)

	assert.NoError(t, err)
	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestGrantPermissionToRole_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "non_existent_role", "/api/v1/articles", "POST"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	deps.Enforcer.AssertNotCalled(t, "AddPolicy", mock.Anything, mock.Anything, mock.Anything)
}

func TestGrantPermissionToRole_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "POST"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, errors.New("db error"))

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestGrantPermissionToRole_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "POST"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", role, "global", path, method).Return(false, errors.New("casbin error"))

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)
	assert.Error(t, err)
	assert.Equal(t, errors.New("casbin error"), err)
}

func TestGrantPermissionToRole_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	assert.Error(t, uc.GrantPermissionToRole(context.Background(), "", "path", "GET"))
	assert.Error(t, uc.GrantPermissionToRole(context.Background(), "role", "", "GET"))
	assert.Error(t, uc.GrantPermissionToRole(context.Background(), "role", "path", ""))
}

func TestRevokePermissionFromRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "editor", "/api/v1/articles", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", role, "global", path, method).Return(true, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)

	assert.NoError(t, err)
	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestRevokePermissionFromRole_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "non_existent_role", "/api/v1/articles", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	deps.Enforcer.AssertNotCalled(t, "RemovePolicy", mock.Anything, mock.Anything, mock.Anything)
}

func TestRevokePermissionFromRole_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, errors.New("db error"))

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestRevokePermissionFromRole_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", role, "global", path, method).Return(false, errors.New("casbin error"))

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)
	assert.Error(t, err)
	assert.Equal(t, errors.New("casbin error"), err)
}

func TestRevokePermissionFromRole_PolicyNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	role, path, method := "role", "/path", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", role, "global", path, method).Return(false, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)
	assert.Error(t, err)
	assert.Equal(t, errors.New("policy to revoke not found in domain global"), err)
}

func TestRevokePermissionFromRole_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	assert.Error(t, uc.RevokePermissionFromRole(context.Background(), "", "path", "GET"))
	assert.Error(t, uc.RevokePermissionFromRole(context.Background(), "role", "", "GET"))
	assert.Error(t, uc.RevokePermissionFromRole(context.Background(), "role", "path", ""))
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

	deps.Enforcer.On("GetFilteredPolicy", 0, role).Return(expectedPolicies, nil)

	policies, err := uc.GetPermissionsForRole(context.Background(), role)
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicies, policies)
}

func TestGetPermissionsForRole_Error(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	deps.Enforcer.On("GetFilteredPolicy", 0, role).Return(nil, errors.New("casbin error"))

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
