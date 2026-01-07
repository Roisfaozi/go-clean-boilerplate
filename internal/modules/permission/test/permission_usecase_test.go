package test

import (
	"context"
	"io"
	"testing"

	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
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
	Enforcer *permissionMocks.IEnforcer
	RoleRepo *roleMocks.MockRoleRepository
	UserRepo *userMocks.MockUserRepository
}

func setupPermissionTest() (*permissionTestDeps, usecase.IPermissionUseCase) {
	deps := &permissionTestDeps{
		Enforcer: new(permissionMocks.IEnforcer),
		RoleRepo: new(roleMocks.MockRoleRepository),
		UserRepo: new(userMocks.MockUserRepository),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewPermissionUseCase(deps.Enforcer, log, deps.RoleRepo, deps.UserRepo)
	return deps, uc
}

func TestPermissionUseCase_AssignRoleToUser_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID := "u1"
	role := "admin"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID).Return(true, nil)
	deps.Enforcer.On("AddGroupingPolicy", userID, role).Return(true, nil)

	err := uc.AssignRoleToUser(context.Background(), userID, role)

	assert.NoError(t, err)
	deps.Enforcer.AssertExpectations(t)
}

func TestPermissionUseCase_AssignRoleToUser_UserNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID := "unknown"
	role := "admin"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, role)

	assert.ErrorIs(t, err, exception.ErrNotFound)
	deps.RoleRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
}

func TestPermissionUseCase_AssignRoleToUser_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID := "u1"
	role := "unknown"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, role)

	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestPermissionUseCase_GrantPermissionToRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	path := "/api/v1/resource"
	method := "GET"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", role, path, method).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)

	assert.NoError(t, err)
}

func TestPermissionUseCase_GrantPermissionToRole_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "unknown"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.GrantPermissionToRole(context.Background(), role, "/path", "GET")

	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestPermissionUseCase_RevokePermissionFromRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	path := "/api/v1/resource"
	method := "GET"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", role, path, method).Return(true, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)

	assert.NoError(t, err)
}

func TestPermissionUseCase_RevokePermissionFromRole_PolicyNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	path := "/api/v1/resource"
	method := "GET"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", role, path, method).Return(false, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)

	assert.ErrorContains(t, err, "policy to revoke not found")
}

func TestPermissionUseCase_GetAllPermissions_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	policies := [][]string{{"admin", "/path", "GET"}}

	deps.Enforcer.On("GetPolicy").Return(policies, nil)

	res, err := uc.GetAllPermissions(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, policies, res)
}

func TestPermissionUseCase_GetPermissionsForRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	policies := [][]string{{"admin", "/path", "GET"}}

	deps.Enforcer.On("GetFilteredPolicy", 0, role).Return(policies, nil)

	res, err := uc.GetPermissionsForRole(context.Background(), role)

	assert.NoError(t, err)
	assert.Equal(t, policies, res)
}

func TestPermissionUseCase_UpdatePermission_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"admin", "/path", "GET"}
	newP := []string{"admin", "/path", "POST"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(true, nil)

	success, err := uc.UpdatePermission(context.Background(), oldP, newP)

	assert.NoError(t, err)
	assert.True(t, success)
}

func TestPermissionUseCase_UpdatePermission_InvalidInput(t *testing.T) {
	_, uc := setupPermissionTest()
	_, err := uc.UpdatePermission(context.Background(), nil, nil)
	assert.ErrorContains(t, err, "cannot be empty")
}
