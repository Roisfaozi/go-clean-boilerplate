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

func setupRoleUsecaseTest() (usecase.IPermissionUseCase, *mocks.IEnforcer, *roleMocks.MockRoleRepository, *userMocks.MockUserRepository) {
	mockEnforcer := new(mocks.IEnforcer)
	mockRoleRepo := new(roleMocks.MockRoleRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	uc := usecase.NewPermissionUseCase(mockEnforcer, logrus.New(), mockRoleRepo, mockUserRepo)
	return uc, mockEnforcer, mockRoleRepo, mockUserRepo
}

func TestAssignRoleToUser_Success(t *testing.T) {

	uc, mockEnforcer, mockRoleRepo, mockUserRepo := setupRoleUsecaseTest()

	userID, roleName := "user123", "editor"

	// Mock UserRepo
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)

	// Mock RoleRepo
	mockRoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)

	mockEnforcer.On("RemoveFilteredGroupingPolicy", 0, userID).Return(true, nil)
	mockEnforcer.On("AddGroupingPolicy", userID, roleName).Return(true, nil)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockEnforcer.AssertExpectations(t)
}

func TestAssignRoleToUser_UserNotFound(t *testing.T) {
	uc, mockEnforcer, mockRoleRepo, mockUserRepo := setupRoleUsecaseTest()

	userID, roleName := "user123", "editor"

	// Mock UserRepo Fail
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)

	mockRoleRepo.AssertNotCalled(t, "FindByName", mock.Anything, mock.Anything)
	mockEnforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything)
}

func TestAssignRoleToUser_RoleNotFound(t *testing.T) {
	uc, mockEnforcer, mockRoleRepo, mockUserRepo := setupRoleUsecaseTest()

	userID, roleName := "user123", "non_existent_role"

	// Mock UserRepo Success
	mockUserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)

	// Mock RoleRepo Fail
	mockRoleRepo.On("FindByName", mock.Anything, roleName).Return(nil, gorm.ErrRecordNotFound)

	err := uc.AssignRoleToUser(context.Background(), userID, roleName)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	mockEnforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything)
}

func TestGrantPermissionToRole_Success(t *testing.T) {
	uc, mockEnforcer, mockRoleRepo, _ := setupRoleUsecaseTest()

	role, path, method := "editor", "/api/v1/articles", "POST"
	mockRoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	mockEnforcer.On("AddPolicy", role, path, method).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)

	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockEnforcer.AssertExpectations(t)
}

func TestGrantPermissionToRole_RoleNotFound(t *testing.T) {
	uc, mockEnforcer, mockRoleRepo, _ := setupRoleUsecaseTest()

	role, path, method := "non_existent_role", "/api/v1/articles", "POST"
	mockRoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.GrantPermissionToRole(context.Background(), role, path, method)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	mockEnforcer.AssertNotCalled(t, "AddPolicy", mock.Anything, mock.Anything, mock.Anything)
}

func TestRevokePermissionFromRole_Success(t *testing.T) {
	uc, mockEnforcer, mockRoleRepo, _ := setupRoleUsecaseTest()

	role, path, method := "editor", "/api/v1/articles", "DELETE"
	mockRoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	mockEnforcer.On("RemovePolicy", role, path, method).Return(true, nil)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)

	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockEnforcer.AssertExpectations(t)
}

func TestRevokePermissionFromRole_RoleNotFound(t *testing.T) {
	uc, mockEnforcer, mockRoleRepo, _ := setupRoleUsecaseTest()

	role, path, method := "non_existent_role", "/api/v1/articles", "DELETE"
	mockRoleRepo.On("FindByName", mock.Anything, role).Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokePermissionFromRole(context.Background(), role, path, method)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrBadRequest, err)
	mockEnforcer.AssertNotCalled(t, "RemovePolicy", mock.Anything, mock.Anything, mock.Anything)
}
