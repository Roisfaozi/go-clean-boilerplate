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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupRoleTest() (*mocks.MockRoleRepository, *mocking.MockWithTransactionManager, usecase.RoleUseCase) {
	mockRepo := new(mocks.MockRoleRepository)
	mockTM := new(mocking.MockWithTransactionManager)
	uc := usecase.NewRoleUseCase(logrus.New(), mockTM, mockRepo)
	return mockRepo, mockTM, uc
}

func TestRoleUseCase_Create(t *testing.T) {
	t.Run("Success - Basic Role Creation", func(t *testing.T) {
		mockRepo, mockTM, uc := setupRoleTest()
		req := &model.CreateRoleRequest{Name: "new_role", Description: "A new role for testing"}

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		mockRepo.On("FindByName", mock.Anything, "new_role").Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *entity.Role) bool {
			return r.Name == "new_role" && r.Description == "A new role for testing"
		})).Return(nil)

		res, err := uc.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Name, res.Name)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Role Already Exists", func(t *testing.T) {
		mockRepo, mockTM, uc := setupRoleTest()
		req := &model.CreateRoleRequest{Name: "existing_role"}

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Return(exception.ErrConflict).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			})
		mockRepo.On("FindByName", mock.Anything, "existing_role").
			Return(&entity.Role{Name: "existing_role"}, nil)

		res, err := uc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, exception.ErrConflict)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error on Create", func(t *testing.T) {
		mockRepo, mockTM, uc := setupRoleTest()
		req := &model.CreateRoleRequest{Name: "create_error_role"}

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Return(exception.ErrInternalServer).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			})
		mockRepo.On("FindByName", mock.Anything, "create_error_role").
			Return((*entity.Role)(nil), gorm.ErrRecordNotFound)
		mockRepo.On("Create", mock.Anything, mock.Anything).
			Return(errors.New("database error"))

		res, err := uc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestRoleUseCase_GetAll(t *testing.T) {
	t.Run("Success - Get All Roles", func(t *testing.T) {
		mockRepo, mockTM, uc := setupRoleTest()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			})

		mockRepo.On("FindAll", mock.Anything).
			Return([]*entity.Role{
				{ID: "1", Name: "admin"},
				{ID: "2", Name: "user"},
			}, nil)

		result, err := uc.GetAll(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, mockTM, uc := setupRoleTest()

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Return(exception.ErrInternalServer).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			})

		mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("database error"))

		result, err := uc.GetAll(context.Background())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestRoleUseCase_Delete(t *testing.T) {

	roleID := "role-123"

	t.Run("Success - Role Deleted", func(t *testing.T) {

		mockRepo, mockTM, uc := setupRoleTest()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {

				fn := args.Get(1).(func(context.Context) error)

				_ = fn(context.Background())

			})

		mockRepo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "editor"}, nil)

		mockRepo.On("Delete", mock.Anything, roleID).Return(nil)

		err := uc.Delete(context.Background(), roleID)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)

		mockTM.AssertExpectations(t)

	})

	t.Run("Error - Role Not Found", func(t *testing.T) {

		mockRepo, mockTM, uc := setupRoleTest()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrNotFound).
			Run(func(args mock.Arguments) {

				fn := args.Get(1).(func(context.Context) error)

				err := fn(context.Background())

				assert.Equal(t, exception.ErrNotFound, err)

			})

		mockRepo.On("FindByID", mock.Anything, roleID).Return(nil, gorm.ErrRecordNotFound)

		err := uc.Delete(context.Background(), roleID)

		assert.ErrorIs(t, err, exception.ErrNotFound)

		mockRepo.AssertExpectations(t)

		mockTM.AssertExpectations(t)

	})

	t.Run("Error - Cannot Delete Superadmin", func(t *testing.T) {

		mockRepo, mockTM, uc := setupRoleTest()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrForbidden).
			Run(func(args mock.Arguments) {

				fn := args.Get(1).(func(context.Context) error)

				err := fn(context.Background())

				assert.Equal(t, exception.ErrForbidden, err)

			})

		mockRepo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "role:superadmin"}, nil)

		err := uc.Delete(context.Background(), roleID)

		assert.ErrorIs(t, err, exception.ErrForbidden)

		mockRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)

		mockRepo.AssertExpectations(t)

		mockTM.AssertExpectations(t)

	})

	t.Run("Error - Database Error During Delete", func(t *testing.T) {

		mockRepo, mockTM, uc := setupRoleTest()

		dbError := errors.New("database error")

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrInternalServer).
			Run(func(args mock.Arguments) {

				fn := args.Get(1).(func(context.Context) error)

				err := fn(context.Background())

				assert.Equal(t, exception.ErrInternalServer, err)

			})

		mockRepo.On("FindByID", mock.Anything, roleID).Return(&entity.Role{ID: roleID, Name: "editor"}, nil)

		mockRepo.On("Delete", mock.Anything, roleID).Return(dbError)

		err := uc.Delete(context.Background(), roleID)

		assert.ErrorIs(t, err, exception.ErrInternalServer)

		mockRepo.AssertExpectations(t)

		mockTM.AssertExpectations(t)

	})

}

func TestRoleUseCase_GetAllRolesDynamic(t *testing.T) {

	t.Run("Success - With Dynamic Filter", func(t *testing.T) {

		mockRepo, mockTM, uc := setupRoleTest()

		mockRoles := []*entity.Role{

			{ID: "role1", Name: "Dynamic Role 1"},

			{ID: "role2", Name: "Dynamic Role 2"},
		}

		filter := &querybuilder.DynamicFilter{

			Filter: map[string]querybuilder.Filter{

				"Name": {Type: "contains", From: "Dynamic"},
			},
		}

		mockRepo.On("FindAllDynamic", mock.Anything, filter).Return(mockRoles, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {

				fn := args.Get(1).(func(context.Context) error)

				_ = fn(context.Background())

			}).Return(nil)

		result, err := uc.GetAllRolesDynamic(context.Background(), filter)

		assert.NoError(t, err)

		assert.Len(t, result, 2)

		assert.Equal(t, "role1", result[0].ID)

		assert.Equal(t, "Dynamic Role 1", result[0].Name)

		mockRepo.AssertExpectations(t)

		mockTM.AssertExpectations(t)

	})

	t.Run("Error - Database Error", func(t *testing.T) {

		mockRepo, mockTM, uc := setupRoleTest()

		dbError := errors.New("database error")

		expectedError := exception.ErrInternalServer

		filter := &querybuilder.DynamicFilter{

			Filter: map[string]querybuilder.Filter{

				"Name": {Type: "contains", From: "Error"},
			},
		}

		mockRepo.On("FindAllDynamic", mock.Anything, filter).Return(nil, dbError)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {

				fn := args.Get(1).(func(context.Context) error)

				err := fn(context.Background())

				assert.Equal(t, expectedError, err)

			}).Return(expectedError)

		result, err := uc.GetAllRolesDynamic(context.Background(), filter)

		assert.Error(t, err)

		assert.Nil(t, result)

		assert.Equal(t, expectedError, err)

		mockRepo.AssertExpectations(t)

		mockTM.AssertExpectations(t)

	})

}
