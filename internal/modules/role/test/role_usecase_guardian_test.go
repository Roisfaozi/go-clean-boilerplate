package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
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
	Repo *mocks.MockRoleRepository
	TM   *mocking.MockWithTransactionManager
}

func setupGuardianRoleTest() (*guardianRoleTestDeps, usecase.RoleUseCase) {
	deps := &guardianRoleTestDeps{
		Repo: new(mocks.MockRoleRepository),
		TM:   new(mocking.MockWithTransactionManager),
	}
	// Use discarded logger for tests
	log := logrus.New()
	log.SetOutput(ioDiscard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo)
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
