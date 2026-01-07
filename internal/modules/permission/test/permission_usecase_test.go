package test

import (
	"context"
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

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID).Return(true, nil)
	deps.Enforcer.On("AddGroupingPolicy", userID, roleName).Return(true, nil)

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

func TestGrantPermissionToRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "editor", "/api/v1/articles", "POST"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", role, path, method).Return(true, nil)

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

func TestRevokePermissionFromRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()

	role, path, method := "editor", "/api/v1/articles", "DELETE"
	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("RemovePolicy", role, path, method).Return(true, nil)

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
