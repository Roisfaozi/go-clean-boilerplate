package test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type permTestDeps struct {
	Enforcer *permMocks.IEnforcer
	RoleRepo *roleMocks.MockRoleRepository
	UserRepo *userMocks.MockUserRepository
}

func setupPermissionExtendedTest() (*permTestDeps, usecase.IPermissionUseCase) {
	deps := &permTestDeps{
		Enforcer: new(permMocks.IEnforcer),
		RoleRepo: new(roleMocks.MockRoleRepository),
		UserRepo: new(userMocks.MockUserRepository),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewPermissionUseCase(deps.Enforcer, log, deps.RoleRepo, deps.UserRepo)
	return deps, uc
}

func TestAssignRoleToUser_Extended(t *testing.T) {
	deps, uc := setupPermissionExtendedTest()

	t.Run("User_Not_Found", func(t *testing.T) {
		deps.UserRepo.On("FindByID", mock.Anything, "user1").Return(nil, gorm.ErrRecordNotFound).Once()
		err := uc.AssignRoleToUser(context.Background(), "user1", "role1", "global")
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Role_Not_Found", func(t *testing.T) {
		deps.UserRepo.On("FindByID", mock.Anything, "user1").Return(nil, nil).Once()
		deps.RoleRepo.On("FindByName", mock.Anything, "role1").Return(nil, gorm.ErrRecordNotFound).Once()
		err := uc.AssignRoleToUser(context.Background(), "user1", "role1", "global")
		assert.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("Remove_Existing_Error", func(t *testing.T) {
		deps.UserRepo.On("FindByID", mock.Anything, "user1").Return(nil, nil).Once()
		deps.RoleRepo.On("FindByName", mock.Anything, "role1").Return(nil, nil).Once()

		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, "user1", "", "global").Return(false, errors.New("remove error")).Once()

		err := uc.AssignRoleToUser(context.Background(), "user1", "role1", "global")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestGrantPermissionToRole_Extended(t *testing.T) {
	deps, uc := setupPermissionExtendedTest()

	t.Run("AddPolicy_Error", func(t *testing.T) {
		deps.RoleRepo.On("FindByName", mock.Anything, "role1").Return(nil, nil).Once()
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddPolicy", "role1", "global", "/path", "GET").Return(false, errors.New("add error")).Once()

		err := uc.GrantPermissionToRole(context.Background(), "role1", "/path", "GET", "global")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "add error")
	})
}

func TestBatchCheckPermission_Extended(t *testing.T) {
	deps, uc := setupPermissionExtendedTest()

	t.Run("Partial_Failure", func(t *testing.T) {
		items := []model.PermissionCheckItem{
			{Resource: "res1", Action: "act1"},
			{Resource: "res2", Action: "act2"}, // Will fail
		}

		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("Enforce", "user1", "global", "res1", "act1").Return(true, nil).Once()
		deps.Enforcer.On("Enforce", "user1", "global", "res2", "act2").Return(false, errors.New("enforce fail")).Once()

		results, err := uc.BatchCheckPermission(context.Background(), "user1", items)
		assert.NoError(t, err)
		assert.True(t, results["res1:act1"])
		assert.False(t, results["res2:act2"])
	})
}
