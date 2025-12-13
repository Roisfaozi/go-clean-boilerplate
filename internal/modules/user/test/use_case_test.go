package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupUserTest() (*mocks.MockUserRepository, *mocking.MockWithTransactionManager, *permMocks.IEnforcer, usecase.UserUseCase) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockWithTransactionManager)
	mockEnforcer := new(permMocks.IEnforcer)
	uc := usecase.NewUserUseCase(logrus.New(), mockTM, mockRepo, mockEnforcer)
	return mockRepo, mockTM, mockEnforcer, uc
}

func TestUserUseCase_Create_Success(t *testing.T) {
	mockRepo, mockTM, mockEnforcer, uc := setupUserTest()

	testReq := &model.RegisterUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	mockRepo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	mockEnforcer.On("AddGroupingPolicy", mock.AnythingOfType("string"), "role:user").Return(true, nil)

	mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).
		Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
	mockTM.AssertExpectations(t)
	mockEnforcer.AssertExpectations(t)
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		expectedUser := &entity.User{ID: "test123", Name: "Test User"}

		mockRepo.On("FindByID", mock.Anything, "test123").Return(expectedUser, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.GetUserByID(context.Background(), "test123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test123", result.ID)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			}).Return(exception.ErrNotFound)

		result, err := uc.GetUserByID(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"

		mockRepo.On("FindByID", mock.Anything, sqlInjectionID).Return(nil, gorm.ErrInvalidData)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Return(exception.ErrInternalServer)

		result, err := uc.GetUserByID(context.Background(), sqlInjectionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		dbError := errors.New("database connection failed")
		expectedError := exception.ErrInternalServer

		mockRepo.On("FindByID", mock.Anything, "db-error").Return(nil, dbError)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, expectedError, err)
			}).Return(expectedError)

		result, err := uc.GetUserByID(context.Background(), "db-error")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_GetAllUsers(t *testing.T) {
	t.Run("Success - With Users", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		mockUsers := []*entity.User{
			{ID: "user1", Name: "User One"},
			{ID: "user2", Name: "User Two"},
		}
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		mockRepo.On("FindAll", mock.Anything, req).Return(mockUsers, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		req := &model.GetUserListRequest{Page: 1, Limit: 10}
		mockRepo.On("FindAll", mock.Anything, req).Return([]*entity.User{}, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, result)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		dbError := errors.New("database connection failed")
		expectedError := exception.ErrInternalServer
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		mockRepo.On("FindAll", mock.Anything, req).Return(nil, dbError)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, expectedError, err)
			}).Return(expectedError)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_Current(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		expectedUser := &entity.User{ID: "current-user", Name: "Current User"}
		testReq := &model.GetUserRequest{ID: "current-user"}

		mockRepo.On("FindByID", mock.Anything, "current-user").Return(expectedUser, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.Current(context.Background(), testReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "current-user", result.ID)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		testReq := &model.GetUserRequest{ID: "nonexistent"}

		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrNotFound)
			}).Return(exception.ErrNotFound)

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		testReq := &model.GetUserRequest{ID: "db-error"}
		dbError := errors.New("database error")

		mockRepo.On("FindByID", mock.Anything, "db-error").Return(nil, dbError)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			}).Return(exception.ErrInternalServer)

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_Update(t *testing.T) {
	t.Run("Success - User Updated", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID:   "user123",
			Name: "Updated User",
		}

		existingUser := &entity.User{
			ID:   "user123",
			Name: "Original User",
		}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == "user123" && u.Name == "Updated User"
		})).Return(nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.Update(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user123", result.ID)
		assert.Equal(t, "Updated User", result.Name)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		updateReq := &model.UpdateUserRequest{
			ID:   "nonexistent",
			Name: "New Name",
		}

		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.ErrorIs(t, err, exception.ErrNotFound)
			}).Return(exception.ErrNotFound)

		result, err := uc.Update(context.Background(), updateReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	userID := "user-to-delete"

	t.Run("Success - User Deleted", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		mockRepo.On("Delete", mock.Anything, userID).Return(nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.NoError(t, err)
			}).Return(nil)

		err := uc.DeleteUser(context.Background(), userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			}).Return(exception.ErrNotFound)

		err := uc.DeleteUser(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"

		mockRepo.On("FindByID", mock.Anything, sqlInjectionID).Return(nil, gorm.ErrInvalidData)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Return(exception.ErrBadRequest)

		err := uc.DeleteUser(context.Background(), sqlInjectionID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error During Delete", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		dbError := errors.New("database error during delete")

		mockRepo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		mockRepo.On("Delete", mock.Anything, userID).Return(dbError)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Return(dbError)

		err := uc.DeleteUser(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Context Canceled", func(t *testing.T) {
		_, mockTM, _, uc := setupUserTest()
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

func TestUserUseCase_GetAllUsersDynamic(t *testing.T) {
	t.Run("Success - With Dynamic Filter", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
		mockUsers := []*entity.User{
			{ID: "user1", Name: "Dynamic User 1"},
			{ID: "user2", Name: "Dynamic User 2"},
		}

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Dynamic"},
			},
		}

		mockRepo.On("FindAllDynamic", mock.Anything, filter).Return(mockUsers, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "Dynamic User 1", result[0].Name)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, mockTM, _, uc := setupUserTest()
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

		result, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}
