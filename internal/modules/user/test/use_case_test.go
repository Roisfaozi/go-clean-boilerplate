package test

import (
	"context"
	"errors"
	"io"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
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

func setupUserTest() (*mocks.MockUserRepository, *mocking.MockWithTransactionManager, *permMocks.IEnforcer, *auditMocks.MockAuditUseCase, *authMocks.MockAuthUseCase, usecase.UserUseCase) {
	mockRepo := new(mocks.MockUserRepository)
	mockTM := new(mocking.MockWithTransactionManager)
	mockEnforcer := new(permMocks.IEnforcer)
	mockAuditUC := new(auditMocks.MockAuditUseCase)
	mockAuthUC := new(authMocks.MockAuthUseCase)

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewUserUseCase(mockTM, log, mockRepo, mockEnforcer, mockAuditUC, mockAuthUC)

	return mockRepo, mockTM, mockEnforcer, mockAuditUC, mockAuthUC, uc
}

func TestUserUseCase_Create_Success(t *testing.T) {
	mockRepo, _, mockEnforcer, mockAuditUC, _, uc := setupUserTest()

	testReq := &model.RegisterUserRequest{
		Username: "testuser", Email: "test@example.com", Name: "Test User", Password: "password123",
	}

	mockRepo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	mockEnforcer.On("AddGroupingPolicy", mock.AnythingOfType("string"), "role:user").Return(true, nil)
	mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
	mockEnforcer.AssertExpectations(t)
	mockAuditUC.AssertExpectations(t)
}

func TestUserUseCase_Create_Conflict(t *testing.T) {
	mockRepo, _, _, _, _, uc := setupUserTest()

	t.Run("Username Exists", func(t *testing.T) {
		req := &model.RegisterUserRequest{
			Username: "existing", Email: "new@example.com", Password: "password123", Name: "Test",
		}

		mockRepo.On("FindByUsername", mock.Anything, "existing").Return(&entity.User{Username: "existing"}, nil)

		_, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrConflict)
	})

	t.Run("Email Exists", func(t *testing.T) {
		req := &model.RegisterUserRequest{
			Username: "newuser", Email: "existing@example.com", Password: "password123", Name: "Test",
		}

		mockRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(&entity.User{Email: "existing@example.com"}, nil)

		_, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrConflict)
	})
}

func TestUserUseCase_Create_RepoError(t *testing.T) {
	mockRepo, _, _, _, _, uc := setupUserTest()
	req := &model.RegisterUserRequest{
		Username: "user", Email: "test@example.com", Password: "password123", Name: "Test",
	}

	mockRepo.On("FindByUsername", mock.Anything, "user").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := uc.Create(context.Background(), req)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		expectedUser := &entity.User{ID: "test123", Name: "Test User"}

		mockRepo.On("FindByID", mock.Anything, "test123").Return(expectedUser, nil)

		result, err := uc.GetUserByID(context.Background(), "test123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test123", result.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))

		result, err := uc.GetUserByID(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		_, _, _, _, _, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"

		result, err := uc.GetUserByID(context.Background(), sqlInjectionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrBadRequest, err)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		dbError := errors.New("database connection failed")
		
		mockRepo.On("FindByID", mock.Anything, "db-error").Return(nil, dbError)

		result, err := uc.GetUserByID(context.Background(), "db-error")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_GetAllUsers(t *testing.T) {
	t.Run("Success - With Users", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		mockUsers := []*entity.User{
			{ID: "user1", Name: "User One"},
			{ID: "user2", Name: "User Two"},
		}
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		mockRepo.On("FindAll", mock.Anything, req).Return(mockUsers, nil)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		req := &model.GetUserListRequest{Page: 1, Limit: 10}
		mockRepo.On("FindAll", mock.Anything, req).Return([]*entity.User{}, nil)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		dbError := errors.New("database connection failed")
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		mockRepo.On("FindAll", mock.Anything, req).Return(nil, dbError)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrInternalServer, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_Current(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		expectedUser := &entity.User{ID: "current-user", Name: "Current User"}
		testReq := &model.GetUserRequest{ID: "current-user"}

		mockRepo.On("FindByID", mock.Anything, "current-user").Return(expectedUser, nil)

		result, err := uc.Current(context.Background(), testReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "current-user", result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		testReq := &model.GetUserRequest{ID: "nonexistent"}

		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_Update(t *testing.T) {
	t.Run("Success - User Updated", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Name: "Updated User",
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

		result, err := uc.Update(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user123", result.ID)
		assert.Equal(t, "Updated User", result.Name)

		mockRepo.AssertExpectations(t)
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Update Password - Success", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Password: "newpassword123",
		}

		existingUser := &entity.User{ID: "user123"}

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		_, err := uc.Update(context.Background(), request)
		assert.NoError(t, err)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		updateReq := &model.UpdateUserRequest{
			ID:   "nonexistent",
			Name: "New Name",
		}

		mockRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		result, err := uc.Update(context.Background(), updateReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Update Fails", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		request := &model.UpdateUserRequest{ID: "user123", Name: "Updated"}
		existingUser := &entity.User{ID: "user123"}
		dbErr := errors.New("db update failed")

		mockRepo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(dbErr)

		_, err := uc.Update(context.Background(), request)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	actorUserID := "admin-user-id"
	cleanID := "019b9150-304e-79d0-aa16-4a2b44347a08"
	deleteReq := &model.DeleteUserRequest{ID: cleanID}

	t.Run("Success - User Deleted", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID, Username: "deletedUser"}, nil)
		mockRepo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, deleteReq.ID).Return(nil, errors.New("user not found"))

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		_, _, _, mockAuditUC, _, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"
		deleteReqSqli := &model.DeleteUserRequest{ID: sqlInjectionID}

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReqSqli)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrBadRequest, err)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Database Error During Delete", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		dbError := errors.New("internal server error")
		
		mockRepo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID}, nil)
		mockRepo.On("Delete", mock.Anything, deleteReq.ID).Return(dbError)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockRepo.AssertExpectations(t)
		mockAuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})
}

func TestUserUseCase_GetAllUsersDynamic(t *testing.T) {
	t.Run("Success - With Dynamic Filter", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
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

		result, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "Dynamic User 1", result[0].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		dbError := errors.New("database error")
		expectedError := exception.ErrInternalServer

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Error"},
			},
		}

		mockRepo.On("FindAllDynamic", mock.Anything, filter).Return(nil, dbError)

		result, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUseCase_UpdateStatus(t *testing.T) {
	t.Run("Success - Active", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, _, uc := setupUserTest()
		userID := "user123"
		status := entity.UserStatusActive

		mockRepo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		mockRepo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
		mockAuditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "UPDATE_STATUS"
		})).Return(nil)

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Success - Banned (Revoke Sessions)", func(t *testing.T) {
		mockRepo, _, _, mockAuditUC, mockAuthUC, uc := setupUserTest()
		userID := "user123"
		status := entity.UserStatusBanned

		mockRepo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		mockRepo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
		
		mockAuthUC.On("RevokeAllSessions", mock.Anything, userID).Return(nil)
		
		mockAuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAuthUC.AssertExpectations(t)
		mockAuditUC.AssertExpectations(t)
	})

	t.Run("Error - Invalid Status", func(t *testing.T) {
		_, _, _, _, _, uc := setupUserTest()
		err := uc.UpdateStatus(context.Background(), "user123", "invalid_status")
		assert.Equal(t, exception.ErrValidationError, err)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		mockRepo, _, _, _, _, uc := setupUserTest()
		mockRepo.On("FindByID", mock.Anything, "unknown").Return(nil, errors.New("user not found"))
		
		err := uc.UpdateStatus(context.Background(), "unknown", entity.UserStatusActive)
		assert.Equal(t, exception.ErrNotFound, err)
	})
}