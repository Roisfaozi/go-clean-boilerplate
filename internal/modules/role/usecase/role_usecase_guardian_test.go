package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"gorm.io/gorm"
	"io"
)

type guardianRoleTestDeps struct {
	Repo           *mocks.MockRoleRepository
	TM             *mocking.MockWithTransactionManager
	PermissionMock *permissionMocks.MockIPermissionUseCase
}

func setupGuardianRoleTest() (*guardianRoleTestDeps, usecase.RoleUseCase) {
	deps := &guardianRoleTestDeps{
		Repo:           new(mocks.MockRoleRepository),
		TM:             new(mocking.MockWithTransactionManager),
		PermissionMock: new(permissionMocks.MockIPermissionUseCase),
	}
	// Use discarded logger for tests
	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, deps.PermissionMock)
	return deps, uc
}

// Simple io.Discard equivalent for logrus

func TestRoleUseCase_Create_Guardian_FindByNameError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "error_role", Description: "Test Role"}

	// Mock Transaction to execute the function
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			// We expect the inner function to return error, so we assert it here or let the transaction return it
			_ = fn(context.Background())
		}).Return(exception.ErrInternalServer)

	// Mock FindByName to return a generic error (not ErrRecordNotFound)
	genericErr := errors.New("connection failed")
	deps.Repo.On("FindByNameInScope", mock.Anything, "error_role", (*string)(nil)).Return((*entity.Role)(nil), genericErr)

	res, err := uc.Create(context.Background(), req)

	// Expect ErrInternalServer because the code wraps generic errors
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.Repo.AssertExpectations(t)
	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_Delete_Guardian_FindByIDError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-error-id"

	// Mock Transaction to execute the function
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(exception.ErrInternalServer)

	// Mock FindByID to return a generic error (not ErrRecordNotFound)
	genericErr := errors.New("connection failed")
	deps.Repo.On("FindByID", mock.Anything, roleID).Return((*entity.Role)(nil), genericErr)

	err := uc.Delete(context.Background(), roleID)

	// Expect ErrInternalServer because the code wraps generic errors
	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.Repo.AssertExpectations(t)
	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_Update_Guardian_FindByIDError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-error-id"
	req := &model.UpdateRoleRequest{Description: "Updated Desc"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	genericErr := errors.New("connection failed")
	deps.Repo.On("FindByID", mock.Anything, roleID).Return((*entity.Role)(nil), genericErr)

	res, err := uc.Update(context.Background(), roleID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.Repo.AssertExpectations(t)
	deps.TM.AssertExpectations(t)
}


func TestRoleUseCase_Delete_Guardian_CleanUpPolicyError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-test-id"
	role := &entity.Role{
		ID:   roleID,
		Name: "test_role",
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByID", mock.Anything, roleID).Return(role, nil)
	deps.Repo.On("Delete", mock.Anything, roleID).Return(nil)

	genericErr := errors.New("cleanup failed")
	deps.PermissionMock.On("DeleteRole", mock.Anything, role.Name).Return(genericErr)

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.Repo.AssertExpectations(t)
	deps.TM.AssertExpectations(t)
	deps.PermissionMock.AssertExpectations(t)
}

func TestRoleUseCase_Create_Guardian_TMError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "tm_error_role", Description: "Test TM Error"}

	// Mock Transaction to return error immediately
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(exception.ErrInternalServer)

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_Update_Guardian_TMError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.UpdateRoleRequest{Description: "Test TM Error"}
	roleID := "id123"

	// Mock Transaction to return error immediately
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(exception.ErrInternalServer)

	res, err := uc.Update(context.Background(), roleID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_GetAll_Guardian_TMError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(exception.ErrInternalServer)

	res, err := uc.GetAll(context.Background())

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_GetAllRolesDynamic_Guardian_TMError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(exception.ErrInternalServer)

	res, err := uc.GetAllRolesDynamic(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_Delete_Guardian_TMError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(exception.ErrInternalServer)

	err := uc.Delete(context.Background(), "id123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)

	deps.TM.AssertExpectations(t)
}

func TestRoleUseCase_Create_Guardian_FindByNameSuccess(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "success_role", Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByNameInScope", mock.Anything, "success_role", (*string)(nil)).Return(&entity.Role{ID: "existing-id", Name: "success_role"}, nil)

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrConflict)

	deps.Repo.AssertExpectations(t)
	deps.TM.AssertExpectations(t)
}
func TestRoleUseCase_CreateForOrganization_Guardian_EmptyOrgID(t *testing.T) {
	_, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "success_role", Description: "Test Role"}

	res, err := uc.CreateForOrganization(context.Background(), "", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRoleUseCase_GetOrganizationRoles_Guardian_EmptyOrgID(t *testing.T) {
	_, uc := setupGuardianRoleTest()

	res, err := uc.GetOrganizationRoles(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRoleUseCase_UpdateForOrganization_Guardian_EmptyIDs(t *testing.T) {
	_, uc := setupGuardianRoleTest()
	req := &model.UpdateRoleRequest{Description: "Test Role"}

	res, err := uc.UpdateForOrganization(context.Background(), "", "roleID", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrBadRequest)

	res, err = uc.UpdateForOrganization(context.Background(), "orgID", "", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRoleUseCase_GetOrganizationRoles_Guardian_Error(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	genericErr := errors.New("find failed")
	deps.Repo.On("FindOrganizationRoles", mock.Anything, "org123").Return(([]*entity.Role)(nil), genericErr)

	res, err := uc.GetOrganizationRoles(context.Background(), "org123")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_GetOrganizationRoles_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	roles := []*entity.Role{
		{ID: "r1", Name: "Role 1"},
	}
	deps.Repo.On("FindOrganizationRoles", mock.Anything, "org123").Return(roles, nil)

	res, err := uc.GetOrganizationRoles(context.Background(), "org123")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
	assert.Equal(t, "Role 1", res[0].Name)
}

func TestRoleUseCase_UpdateForOrganization_Guardian_ErrorFind(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.UpdateRoleRequest{Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	genericErr := errors.New("find failed")
	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return((*entity.Role)(nil), genericErr)

	res, err := uc.UpdateForOrganization(context.Background(), "org123", "role123", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_UpdateForOrganization_Guardian_NotFound(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.UpdateRoleRequest{Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return((*entity.Role)(nil), gorm.ErrRecordNotFound)

	res, err := uc.UpdateForOrganization(context.Background(), "org123", "role123", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRoleUseCase_UpdateForOrganization_Guardian_UpdateError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.UpdateRoleRequest{Description: "Test Role"}
	role := &entity.Role{ID: "role123", Name: "Role 1"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return(role, nil)
	genericErr := errors.New("update failed")
	deps.Repo.On("Update", mock.Anything, role).Return(genericErr)

	res, err := uc.UpdateForOrganization(context.Background(), "org123", "role123", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_UpdateForOrganization_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.UpdateRoleRequest{Description: "Test Role Updated"}
	role := &entity.Role{ID: "role123", Name: "Role 1"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return(role, nil)
	deps.Repo.On("Update", mock.Anything, role).Return(nil)

	res, err := uc.UpdateForOrganization(context.Background(), "org123", "role123", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Role 1", res.Name)
	assert.Equal(t, "Test Role Updated", role.Description)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_NotFound(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return((*entity.Role)(nil), gorm.ErrRecordNotFound)

	err := uc.DeleteForOrganization(context.Background(), "org123", "role123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_ErrorFind(t *testing.T) {
	deps, uc := setupGuardianRoleTest()

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	genericErr := errors.New("find failed")
	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return((*entity.Role)(nil), genericErr)

	err := uc.DeleteForOrganization(context.Background(), "org123", "role123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_ErrorDelete(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	role := &entity.Role{ID: "role123", Name: "Role 1"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return(role, nil)
	genericErr := errors.New("delete failed")
	deps.Repo.On("DeleteInOrg", mock.Anything, "org123", "role123").Return(genericErr)

	err := uc.DeleteForOrganization(context.Background(), "org123", "role123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_Success(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	role := &entity.Role{ID: "role123", Name: "Role 1"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return(role, nil)
	deps.Repo.On("DeleteInOrg", mock.Anything, "org123", "role123").Return(nil)
	deps.PermissionMock.On("DeleteRoleInOrg", mock.Anything, role.Name, "org123").Return(nil)

	err := uc.DeleteForOrganization(context.Background(), "org123", "role123")

	assert.NoError(t, err)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_DeleteNotFound(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	role := &entity.Role{ID: "role123", Name: "Role 1"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, "org123", "role123").Return(role, nil)
	deps.Repo.On("DeleteInOrg", mock.Anything, "org123", "role123").Return(gorm.ErrRecordNotFound)

	err := uc.DeleteForOrganization(context.Background(), "org123", "role123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRoleUseCase_Delete_Guardian_SuperadminBlocked(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-superadmin-id"

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "role:superadmin"}, nil)

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestRoleUseCase_Delete_Guardian_DeleteError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-delete-error-id"

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "normal_role"}, nil)
	genericErr := errors.New("delete failed")
	deps.Repo.On("Delete", mock.Anything, roleID).Return(genericErr)

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_Delete_Guardian_DeleteNotFound(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-delete-notfound-id"

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "normal_role"}, nil)
	deps.Repo.On("Delete", mock.Anything, roleID).Return(gorm.ErrRecordNotFound)

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRoleUseCase_Delete_Guardian_CleanUpPolicyErrorOrg(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	roleID := "role-test-id"
	orgID := "org123"
	role := &entity.Role{
		ID:             roleID,
		Name:           "test_role",
		OrganizationID: &orgID,
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByID", mock.Anything, roleID).Return(role, nil)
	deps.Repo.On("Delete", mock.Anything, roleID).Return(nil)

	genericErr := errors.New("cleanup failed")
	deps.PermissionMock.On("DeleteRoleInOrg", mock.Anything, role.Name, orgID).Return(genericErr)

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

// Add dummy ReloadPolicy method to the permission mock to test the interface cast

type reloaderMock struct {
	*permissionMocks.MockIPermissionUseCase
}

func (m *reloaderMock) ReloadPolicy(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRoleUseCase_Delete_Guardian_ReloadPolicyError(t *testing.T) {
	deps := &guardianRoleTestDeps{
		Repo:           new(mocks.MockRoleRepository),
		TM:             new(mocking.MockWithTransactionManager),
	}
	permMock := new(reloaderMock)
	permMock.MockIPermissionUseCase = new(permissionMocks.MockIPermissionUseCase)
	deps.PermissionMock = permMock.MockIPermissionUseCase

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, permMock)

	roleID := "role-test-id"
	role := &entity.Role{
		ID:   roleID,
		Name: "test_role",
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByID", mock.Anything, roleID).Return(role, nil)
	deps.Repo.On("Delete", mock.Anything, roleID).Return(nil)
	permMock.On("DeleteRole", mock.Anything, role.Name).Return(nil)

	permMock.On("ReloadPolicy", mock.Anything).Return(errors.New("reload failed"))

	err := uc.Delete(context.Background(), roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_ReloadPolicyError2(t *testing.T) {
	deps := &guardianRoleTestDeps{
		Repo:           new(mocks.MockRoleRepository),
		TM:             new(mocking.MockWithTransactionManager),
	}
	permMock := new(reloaderMock)
	permMock.MockIPermissionUseCase = new(permissionMocks.MockIPermissionUseCase)
	deps.PermissionMock = permMock.MockIPermissionUseCase

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, permMock)

	roleID := "role-test-id"
	orgID := "org123"
	role := &entity.Role{
		ID:   roleID,
		Name: "test_role",
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, orgID, roleID).Return(role, nil)
	deps.Repo.On("DeleteInOrg", mock.Anything, orgID, roleID).Return(nil)
	permMock.On("DeleteRoleInOrg", mock.Anything, role.Name, orgID).Return(nil)

	permMock.On("ReloadPolicy", mock.Anything).Return(errors.New("reload failed"))

	err := uc.DeleteForOrganization(context.Background(), orgID, roleID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRoleUseCase_DeleteForOrganization_Guardian_ReloadPolicySuccess(t *testing.T) {
	deps := &guardianRoleTestDeps{
		Repo:           new(mocks.MockRoleRepository),
		TM:             new(mocking.MockWithTransactionManager),
	}
	permMock := new(reloaderMock)
	permMock.MockIPermissionUseCase = new(permissionMocks.MockIPermissionUseCase)
	deps.PermissionMock = permMock.MockIPermissionUseCase

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, permMock)

	roleID := "role-test-id"
	orgID := "org123"
	role := &entity.Role{
		ID:   roleID,
		Name: "test_role",
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindOrganizationRoleByID", mock.Anything, orgID, roleID).Return(role, nil)
	deps.Repo.On("DeleteInOrg", mock.Anything, orgID, roleID).Return(nil)
	permMock.On("DeleteRoleInOrg", mock.Anything, role.Name, orgID).Return(nil)

	permMock.On("ReloadPolicy", mock.Anything).Return(nil)

	err := uc.DeleteForOrganization(context.Background(), orgID, roleID)

	assert.NoError(t, err)
}

func TestRoleUseCase_Create_Guardian_ReservedRole(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "role:admin", Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRoleUseCase_Create_Guardian_CreateError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "create_error_role", Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByNameInScope", mock.Anything, "create_error_role", (*string)(nil)).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
	genericErr := errors.New("create failed")
	deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Role")).Return(genericErr)

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

// This tests the branch in Create if GetOrganizationID returns an ID
func TestRoleUseCase_Create_Guardian_WithContextOrgID(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "org_role", Description: "Test Role"}

	// Set an OrganizationID in context to hit orgID != "" in Create
	ctx := context.WithValue(context.Background(), database.OrganizationIDKey, "org123")

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(txCtx context.Context, fn func(context.Context) error) error {
			return fn(txCtx)
		})

	orgIDStr := "org123"
	deps.Repo.On("FindByNameInScope", mock.Anything, "org_role", &orgIDStr).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Role")).Return(nil)

	res, err := uc.Create(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "org_role", res.Name)
}


func TestRoleUseCase_Create_Guardian_FindByNameOtherError(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "org_role", Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	deps.Repo.On("FindByNameInScope", mock.Anything, "org_role", (*string)(nil)).Return((*entity.Role)(nil), errors.New("some error"))

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}


func TestRoleUseCase_CreateForOrganization_Guardian_ValidID(t *testing.T) {
	deps, uc := setupGuardianRoleTest()
	req := &model.CreateRoleRequest{Name: "org_role_2", Description: "Test Role"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
    orgIDStr := "valid_org"
	deps.Repo.On("FindByNameInScope", mock.Anything, "org_role_2", &orgIDStr).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Role")).Return(nil)

	res, err := uc.CreateForOrganization(context.Background(), "valid_org", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "org_role_2", res.Name)
}
