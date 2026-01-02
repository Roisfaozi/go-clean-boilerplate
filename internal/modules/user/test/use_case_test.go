package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
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

func setupUserTest() (*mocks.MockUserRepository, *mocking.MockWithTransactionManager, *permMocks.IEnforcer, *auditMocks.MockAuditUseCase, usecase.UserUseCase) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockWithTransactionManager)
	mockEnforcer := new(permMocks.IEnforcer)
	mockAuditUC := new(auditMocks.MockAuditUseCase)

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewUserUseCase(log, mockTM, mockRepo, mockEnforcer, mockAuditUC)

	return mockRepo, mockTM, mockEnforcer, mockAuditUC, uc
}

func TestUserUseCase_Create(t *testing.T) {
	t.Run("Success - Valid User", func(t *testing.T) {
		mockRepo, mockTM, mockEnforcer, mockAuditUC, uc := setupUserTest()

		testReq := &model.RegisterUserRequest{
			Username: "testuser", Email: "test@example.com", Name: "Test User", Password: "password123",
			IPAddress: "127.0.0.1", UserAgent: "TestAgent",
		}

		mockRepo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Username == "testuser" && u.Email == "test@example.com"
		})).Return(nil)
		mockEnforcer.On("AddGroupingPolicy", mock.AnythingOfType("string"), "role:user").Return(true, nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).
			Return(nil)

		result, err := uc.Create(context.Background(), testReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "testuser", result.Username)

		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
		mockEnforcer.AssertExpectations(t)
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Negative - Invalid Email Format", func(t *testing.T) {
		_, _, _, _, uc := setupUserTest()
		req := &model.RegisterUserRequest{Username: "test", Email: "invalid-email", Password: "password123"}
		resp, err := uc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("Negative - Password Too Weak", func(t *testing.T) {
		_, _, _, _, uc := setupUserTest()
		req := &model.RegisterUserRequest{Username: "test", Email: "test@example.com", Password: "weak"}
		resp, err := uc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "password too weak")
	})

	t.Run("Negative - Password Too Long", func(t *testing.T) {
		_, _, _, _, uc := setupUserTest()
		longPass := strings.Repeat("a", 73)
		req := &model.RegisterUserRequest{Username: "test", Email: "test@example.com", Password: longPass}
		resp, err := uc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "password too long")
	})

	t.Run("Conflict - Username Exists", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		req := &model.RegisterUserRequest{Username: "existing", Email: "new@example.com", Password: "password123"}

		mockRepo.On("FindByUsername", mock.Anything, "existing").Return(&entity.User{}, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrConflict, err)
			}).Return(exception.ErrConflict)

		resp, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrConflict)
		assert.Nil(t, resp)
	})

	t.Run("Conflict - Email Exists", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		req := &model.RegisterUserRequest{Username: "newuser", Email: "existing@example.com", Password: "password123"}

		mockRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(&entity.User{}, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrConflict, err)
			}).Return(exception.ErrConflict)

		resp, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrConflict)
		assert.Nil(t, resp)
	})

	t.Run("Error - Create Failed", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		req := &model.RegisterUserRequest{Username: "test", Email: "test@example.com", Password: "password123"}

		mockRepo.On("FindByUsername", mock.Anything, "test").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrInternalServer, err)
			}).Return(exception.ErrInternalServer)

		resp, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, resp)
	})
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID:        "user123", Name: "Updated User",
			IPAddress: "127.0.0.1", UserAgent: "TestAgent",
		}

		existingUser := &entity.User{
			ID:   "user123",
			Name: "Original User",
		}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == "user123" && u.Name == "Updated User"
		})).Return(nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

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
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Success - Password Update", func(t *testing.T) {
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID:       "user123", Password: "newPassword123",
			IPAddress: "127.0.0.1", UserAgent: "TestAgent",
		}

		existingUser := &entity.User{ID: "user123", Password: "oldHash"}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			// In real test we can't check hash easily, but we know it changed
			return u.ID == "user123" && u.Password != "oldHash"
		})).Return(nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil)

		result, err := uc.Update(context.Background(), request)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Negative - Password Too Weak During Update", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		request := &model.UpdateUserRequest{ID: "user123", Password: "weak"}

		existingUser := &entity.User{ID: "user123"}
		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.EqualError(t, err, "password too weak")
			}).Return(fmt.Errorf("password too weak"))

		_, err := uc.Update(context.Background(), request)
		assert.ErrorContains(t, err, "password too weak")
	})

	t.Run("Negative - Password Too Long During Update", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		request := &model.UpdateUserRequest{ID: "user123", Password: strings.Repeat("a", 73)}

		existingUser := &entity.User{ID: "user123"}
		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.EqualError(t, err, "password too long")
			}).Return(fmt.Errorf("password too long"))

		_, err := uc.Update(context.Background(), request)
		assert.ErrorContains(t, err, "password too long")
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
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
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Update Failed", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		req := &model.UpdateUserRequest{ID: "user123", Name: "New Name"}
		existingUser := &entity.User{ID: "user123"}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrInternalServer, err)
			}).Return(exception.ErrInternalServer)

		_, err := uc.Update(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Error - Update Concurrency (RecordNotFound during update)", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
		req := &model.UpdateUserRequest{ID: "user123", Name: "New Name"}
		existingUser := &entity.User{ID: "user123"}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(gorm.ErrRecordNotFound)

		mockTM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			}).Return(exception.ErrNotFound)

		_, err := uc.Update(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	actorUserID := "admin-user-id"
	deleteReq := &model.DeleteUserRequest{ID: "user-to-delete", IPAddress: "127.0.0.1", UserAgent: "TestAgent"}

	t.Run("Success - User Deleted", func(t *testing.T) {
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID, Username: "deletedUser"}, nil)
		mockRepo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.NoError(t, err)
			}).Return(nil)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, deleteReq.ID).Return(nil, gorm.ErrRecordNotFound)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Equal(t, exception.ErrNotFound, err)
			}).Return(exception.ErrNotFound)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"
		deleteReq := &model.DeleteUserRequest{ID: sqlInjectionID}

		mockRepo.On("FindByID", mock.Anything, sqlInjectionID).Return(nil, gorm.ErrInvalidData)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Return(exception.ErrInternalServer)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Database Error During Delete", func(t *testing.T) {
		mockRepo, mockTM, _, mockAuditUC, uc := setupUserTest()
		dbError := errors.New("database error during delete")
		deleteReq := &model.DeleteUserRequest{ID: "user-to-delete"}

		mockRepo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID}, nil)
		mockRepo.On("Delete", mock.Anything, deleteReq.ID).Return(dbError)

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				err := fn(context.Background())
				assert.Error(t, err)
			}).Return(exception.ErrInternalServer)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrInternalServer, err)
		mockRepo.AssertExpectations(t)
		mockTM.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Context Canceled", func(t *testing.T) {
		_, mockTM, _, mockAuditUC, uc := setupUserTest()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		deleteReq := &model.DeleteUserRequest{ID: "user-to-delete"}

		mockTM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(context.Canceled).
			Run(func(args mock.Arguments) {
			}).Once()

		err := uc.DeleteUser(ctx, actorUserID, deleteReq)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		mockTM.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})
}

func TestUserUseCase_GetAllUsersDynamic(t *testing.T) {
	t.Run("Success - With Dynamic Filter", func(t *testing.T) {
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
		mockRepo, mockTM, _, _, uc := setupUserTest()
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
