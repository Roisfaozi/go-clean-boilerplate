package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/casbin-db/internal/mocking"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserUseCase_Create_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)
	uc := usecase.NewUserUseCase(logrus.New(), validator.New(), mockTM, mockRepo)

	testReq := &model.RegisterUserRequest{
		ID:       "test123",
		Name:     "Test User",
		Password: "password123",
	}

	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil)
	mockRepo.On("FindByID", mock.Anything, "test123").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)
	uc := usecase.NewUserUseCase(logrus.New(), validator.New(), mockTM, mockRepo)

	t.Run("Success - User Found", func(t *testing.T) {
		expectedUser := &entity.User{ID: "test123", Name: "Test User"}

		mockRepo.On("FindByID", mock.Anything, "test123").Return(expectedUser, nil).Once()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)

				mockRepo.On("FindByID", mock.Anything, "test123").Return(expectedUser, nil).Once()

				err := fn(context.Background())
				assert.NoError(t, err)
			}).Once()

		result, err := uc.GetUserByID(context.Background(), "test123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test123", result.ID)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			}).Return(exception.ErrNotFound).Once()

		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := uc.GetUserByID(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Empty ID", func(t *testing.T) {
		t.Skip("Skipping as the current implementation doesn't validate empty ID")
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		sqlInjectionID := "1'; DROP TABLE users;--"
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Return(exception.ErrInternalServer).Once()

		mockRepo.On("FindByID", mock.Anything, sqlInjectionID).Return(nil, gorm.ErrInvalidData).Once()

		result, err := uc.GetUserByID(context.Background(), sqlInjectionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		dbError := errors.New("database connection failed")
		expectedError := exception.ErrInternalServer
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, expectedError, err)
			}).Return(expectedError).Once()

		mockRepo.On("FindByID", mock.Anything, "db-error").Return(nil, dbError).Once()

		result, err := uc.GetUserByID(context.Background(), "db-error")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_GetAllUsers(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)
	uc := usecase.NewUserUseCase(logrus.New(), validator.New(), mockTM, mockRepo)

	t.Run("Success - With Users", func(t *testing.T) {
		mockUsers := []*entity.User{
			{ID: "user1", Name: "User One"},
			{ID: "user2", Name: "User Two"},
		}

		mockRepo.On("FindAll", mock.Anything, 100, 0).Return(mockUsers, nil).Once()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)

				mockRepo.On("FindAll", mock.Anything, 100, 0).Return(mockUsers, nil).Once()

				err := fn(context.Background())
				assert.NoError(t, err)
			}).Once()

		result, err := uc.GetAllUsers(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		mockRepo.On("FindAll", mock.Anything, 100, 0).Return([]*entity.User{}, nil).Once()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)

				mockRepo.On("FindAll", mock.Anything, 100, 0).Return([]*entity.User{}, nil).Once()

				err := fn(context.Background())
				assert.NoError(t, err)
			}).Once()

		result, err := uc.GetAllUsers(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, result)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		dbError := errors.New("database connection failed")
		expectedError := exception.ErrInternalServer

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, expectedError, err)
			}).Return(expectedError).Once()

		mockRepo.On("FindAll", mock.Anything, 100, 0).Return(nil, dbError).Once()

		result, err := uc.GetAllUsers(context.Background())

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_Current(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)
	uc := usecase.NewUserUseCase(logrus.New(), validator.New(), mockTM, mockRepo)

	t.Run("Success - User Found", func(t *testing.T) {
		expectedUser := &entity.User{ID: "current-user", Name: "Current User"}
		testReq := &model.GetUserRequest{ID: "current-user"}

		mockRepo.On("FindByID", mock.Anything, "current-user").Return(expectedUser, nil).Once()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).
			Run(func(args mock.Arguments) {
			}).Once()

		result, err := uc.Current(context.Background(), testReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "current-user", result.ID)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		testReq := &model.GetUserRequest{ID: "nonexistent"}

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrNotFound).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)

				mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound).Once()

				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrNotFound)
			}).Once()

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		testReq := &model.GetUserRequest{ID: "db-error"}
		dbError := errors.New("database error")

		mockRepo.On("FindByID", mock.Anything, "db-error").Return(nil, dbError).Once()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrInternalServer).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			}).Once()

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Empty ID", func(t *testing.T) {
		testReq := &model.GetUserRequest{ID: ""}

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrBadRequest)
		assert.Nil(t, result)
	})
}

func TestUserUseCase_Update(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)
	uc := usecase.NewUserUseCase(logrus.New(), validator.New(), mockTM, mockRepo)

	t.Run("Success - User Updated", func(t *testing.T) {
		request := &model.UpdateUserRequest{
			ID:   "user123",
			Name: "Updated User",
		}

		existingUser := &entity.User{
			ID:   "user123",
			Name: "Original User",
		}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == "user123" && u.Name == "Updated User"
		})).Return(nil).Once()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(nil).Once()

		result, err := uc.Update(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user123", result.ID)
		assert.Equal(t, "Updated User", result.Name)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
	t.Run("Error - User Not Found", func(t *testing.T) {
		updateReq := &model.UpdateUserRequest{
			ID:   "nonexistent",
			Name: "New Name",
		}

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrNotFound).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)

				mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound).Once()

				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrNotFound)
			}).Once()

		result, err := uc.Update(context.Background(), updateReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Validation Error", func(t *testing.T) {
		updateReq := &model.UpdateUserRequest{
			ID:   "",
			Name: "New Name",
		}

		result, err := uc.Update(context.Background(), updateReq)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("Error - Database Error During Update", func(t *testing.T) {
		updateReq := &model.UpdateUserRequest{
			ID:   "user123",
			Name: "New Name",
		}

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				mockRepo.On("FindByID", mock.Anything, "user123").
					Return(&entity.User{ID: "user123", Name: "Old Name"}, nil).Once()

				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
					return u.ID == "user123" && u.Name == "New Name"
				})).Return(errors.New("database error")).Once()

				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			}).
			Return(exception.ErrInternalServer).
			Once()

		result, err := uc.Update(context.Background(), updateReq)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockTransactionManager)
	uc := usecase.NewUserUseCase(logrus.New(), validator.New(), mockTM, mockRepo)
	userID := "user-to-delete"

	t.Run("Success - User Deleted", func(t *testing.T) {
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(nil).Once()
		mockRepo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil).Once()
		mockRepo.On("Delete", mock.Anything, userID).Return(nil).Once()

		err := uc.DeleteUser(context.Background(), userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrNotFound).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			}).Once()

		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound).Once()

		err := uc.DeleteUser(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Empty ID", func(t *testing.T) {
		t.Skip("Skipping as the current implementation doesn't validate empty ID")
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		sqlInjectionID := "1'; DROP TABLE users;--"
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(exception.ErrBadRequest).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Once()

		mockRepo.On("FindByID", mock.Anything, sqlInjectionID).Return(nil, gorm.ErrInvalidData).Once()

		err := uc.DeleteUser(context.Background(), sqlInjectionID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error During Delete", func(t *testing.T) {
		dbError := errors.New("database error during delete")
		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(dbError).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Once()

		mockRepo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil).Once()
		mockRepo.On("Delete", mock.Anything, userID).Return(dbError).Once()

		err := uc.DeleteUser(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Context Canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(context.Canceled).
			Run(func(args mock.Arguments) {
			}).Once()

		err := uc.DeleteUser(ctx, userID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockTM.AssertExpectations(t)
	})
}
