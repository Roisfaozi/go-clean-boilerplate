package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/casbin-db/internal/mocking"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// TestRoleUseCase_Create_Success (Positive Case)
func TestRoleUseCase_Create_Success(t *testing.T) {
	mockRepo := new(mocks.RoleRepository)
	mockTM := new(mocking.MockTransactionManager)

	req := &model.CreateRoleRequest{Name: "new_role", Description: "A new role for testing"}

	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		})

	mockRepo.On("FindByName", mock.Anything, req.Name).Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *entity.Role) bool {
		return r.Name == req.Name
	})).Return(nil)

	uc := usecase.NewRoleUseCase(logrus.New(), validator.New(), mockTM, mockRepo)
	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

// TestRoleUseCase_Create_Conflict (Negative Case: Role Already Exists)
func TestRoleUseCase_Create_Conflict(t *testing.T) {
	mockRepo := new(mocks.RoleRepository)
	mockTM := new(mocking.MockTransactionManager)

	req := &model.CreateRoleRequest{Name: "existing_role"}

	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(nil).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		})

	mockRepo.On("FindByName", mock.Anything, req.Name).Return(&entity.Role{Name: req.Name}, nil)

	uc := usecase.NewRoleUseCase(logrus.New(), validator.New(), mockTM, mockRepo)
	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrConflict, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestRoleUseCase_Create_InvalidPayload (Edge Case: Bad Request)
func TestRoleUseCase_Create_InvalidPayload(t *testing.T) {
	mockRepo := new(mocks.RoleRepository)
	mockTM := new(mocking.MockTransactionManager)

	req := &model.CreateRoleRequest{Name: ""}

	uc := usecase.NewRoleUseCase(logrus.New(), validator.New(), mockTM, mockRepo)
	res, err := uc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrBadRequest, err)
}
