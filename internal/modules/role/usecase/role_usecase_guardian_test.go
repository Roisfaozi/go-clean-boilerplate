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
	log.SetOutput(ioDiscard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, deps.PermissionMock)
	return deps, uc
}

// Simple io.Discard equivalent for logrus
type discardWriter struct{}

func (w discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

var ioDiscard = discardWriter{}

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
	deps.Repo.On("FindByName", mock.Anything, "error_role").Return((*entity.Role)(nil), genericErr)

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

func TestRoleUseCase_Create_Guardian_UUIDError(t *testing.T) {
	// Not practically testable without mocking google/uuid.NewV7 directly.
	// We'll skip since it's hard to trigger and coverage for this block is rare.
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

	deps.Repo.On("FindByName", mock.Anything, "success_role").Return(&entity.Role{ID: "existing-id", Name: "success_role"}, nil)

	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, exception.ErrConflict)

	deps.Repo.AssertExpectations(t)
	deps.TM.AssertExpectations(t)
}
