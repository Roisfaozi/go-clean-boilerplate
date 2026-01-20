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

type guardianPermissionTestDeps struct {
	Enforcer *mocks.IEnforcer
	RoleRepo *roleMocks.MockRoleRepository
	UserRepo *userMocks.MockUserRepository
}

func setupGuardianPermissionTest() (*guardianPermissionTestDeps, usecase.IPermissionUseCase) {
	deps := &guardianPermissionTestDeps{
		Enforcer: new(mocks.IEnforcer),
		RoleRepo: new(roleMocks.MockRoleRepository),
		UserRepo: new(userMocks.MockUserRepository),
	}

	log := logrus.New()
	log.SetOutput(ioDiscard)

	uc := usecase.NewPermissionUseCase(deps.Enforcer, log, deps.RoleRepo, deps.UserRepo)
	return deps, uc
}

// Reuse discardWriter from role tests, but since it is in another package (role/test vs permission/test), we need to redefine it or use standard.
type discardWriter struct{}
func (w discardWriter) Write(p []byte) (n int, err error) { return len(p), nil }
var ioDiscard = discardWriter{}

// ==========================
// RevokeRoleFromUser Tests
// ==========================

func TestRevokeRoleFromUser_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, roleName).Return(true, nil)

	err := uc.RevokeRoleFromUser(context.Background(), userID, roleName)
	assert.NoError(t, err)
}

func TestRevokeRoleFromUser_Guardian_EmptyInput(t *testing.T) {
	_, uc := setupGuardianPermissionTest()
	assert.Error(t, uc.RevokeRoleFromUser(context.Background(), "", "role"))
	assert.Error(t, uc.RevokeRoleFromUser(context.Background(), "user", ""))
}

func TestRevokeRoleFromUser_Guardian_UserNotFound(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRevokeRoleFromUser_Guardian_UserRepoError(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, errors.New("db fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_RoleNotFound(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRevokeRoleFromUser_Guardian_RoleRepoError(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, errors.New("db fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, "u", "r").Return(false, errors.New("casbin fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_RoleNotAssigned(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, "u", "r").Return(false, nil) // Removed = false

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role was not assigned to user")
}

// ==========================
// AddParentRole Tests
// ==========================

func TestAddParentRole_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	child, parent := "editor", "viewer"

	deps.RoleRepo.On("FindByName", mock.Anything, child).Return(&roleEntity.Role{Name: child}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, parent).Return(&roleEntity.Role{Name: parent}, nil)
	deps.Enforcer.On("AddGroupingPolicy", child, parent).Return(true, nil)

	err := uc.AddParentRole(context.Background(), child, parent)
	assert.NoError(t, err)
}

func TestAddParentRole_Guardian_ChildRoleNotFound(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(nil, errors.New("not found"))

	err := uc.AddParentRole(context.Background(), "child", "parent")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestAddParentRole_Guardian_ParentRoleNotFound(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(&roleEntity.Role{Name: "child"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "parent").Return(nil, errors.New("not found"))

	err := uc.AddParentRole(context.Background(), "child", "parent")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestAddParentRole_Guardian_SelfInheritance(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)

	err := uc.AddParentRole(context.Background(), "role", "role")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot inherit from itself")
}

func TestAddParentRole_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	child, parent := "editor", "viewer"

	deps.RoleRepo.On("FindByName", mock.Anything, child).Return(&roleEntity.Role{Name: child}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, parent).Return(&roleEntity.Role{Name: parent}, nil)
	deps.Enforcer.On("AddGroupingPolicy", child, parent).Return(false, errors.New("casbin fail"))

	err := uc.AddParentRole(context.Background(), child, parent)
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

// ==========================
// RemoveParentRole Tests
// ==========================

func TestRemoveParentRole_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, child, parent).Return(true, nil)

	err := uc.RemoveParentRole(context.Background(), child, parent)
	assert.NoError(t, err)
}

func TestRemoveParentRole_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, child, parent).Return(false, errors.New("casbin fail"))

	err := uc.RemoveParentRole(context.Background(), child, parent)
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestRemoveParentRole_Guardian_RelationshipNotFound(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, child, parent).Return(false, nil)

	err := uc.RemoveParentRole(context.Background(), child, parent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inheritance relationship not found")
}

// ==========================
// GetParentRoles Tests
// ==========================

func TestGetParentRoles_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	role := "editor"
	parents := []string{"viewer"}

	deps.Enforcer.On("GetRolesForUser", role).Return(parents, nil)

	res, err := uc.GetParentRoles(context.Background(), role)
	assert.NoError(t, err)
	assert.Equal(t, parents, res)
}

func TestGetParentRoles_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupGuardianPermissionTest()
	role := "editor"

	deps.Enforcer.On("GetRolesForUser", role).Return(nil, errors.New("casbin fail"))

	_, err := uc.GetParentRoles(context.Background(), role)
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}
