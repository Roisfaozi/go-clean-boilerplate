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

type userTestDeps struct {
	Repo     *mocks.MockUserRepository
	TM       *mocking.MockWithTransactionManager
	Enforcer *permMocks.IEnforcer
	AuditUC  *auditMocks.MockAuditUseCase
	AuthUC   *authMocks.MockAuthUseCase
}

func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: new(permMocks.IEnforcer),
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC)

	return deps, uc
}

func TestUserUseCase_Create_Success(t *testing.T) {
	deps, uc := setupUserTest()

	testReq := &model.RegisterUserRequest{
		Username: "testuser", Email: "test@example.com", Name: "Test User", Password: "password123",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.AnythingOfType("string"), "role:user").Return(true, nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	deps.Repo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

func TestUserUseCase_Create_EnforcerError(t *testing.T) {
	deps, uc := setupUserTest()
	testReq := &model.RegisterUserRequest{
		Username: "testuser", Email: "test@example.com", Name: "Test User", Password: "password123",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	// Mock enforcer error - should be logged but not fail the request
	deps.Enforcer.On("AddGroupingPolicy", mock.AnythingOfType("string"), "role:user").Return(false, errors.New("policy error"))
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err) // Should succeed despite role assignment failure (logged only)
	assert.NotNil(t, result)
	deps.Enforcer.AssertExpectations(t)
}


func TestUserUseCase_Create_Conflict(t *testing.T) {
	deps, uc := setupUserTest()

	t.Run("Username Exists", func(t *testing.T) {
		req := &model.RegisterUserRequest{
			Username: "existing", Email: "new@example.com", Password: "password123", Name: "Test",
		}

		deps.Repo.On("FindByUsername", mock.Anything, "existing").Return(&entity.User{Username: "existing"}, nil)

		_, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrConflict)
	})

	t.Run("Email Exists", func(t *testing.T) {
		req := &model.RegisterUserRequest{
			Username: "newuser", Email: "existing@example.com", Password: "password123", Name: "Test",
		}

		deps.Repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
		deps.Repo.On("FindByEmail", mock.Anything, "existing@example.com").Return(&entity.User{Email: "existing@example.com"}, nil)

		_, err := uc.Create(context.Background(), req)
		assert.ErrorIs(t, err, exception.ErrConflict)
	})
}

func TestUserUseCase_Create_RepoError(t *testing.T) {
	deps, uc := setupUserTest()
	req := &model.RegisterUserRequest{
		Username: "user", Email: "test@example.com", Password: "password123", Name: "Test",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "user").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := uc.Create(context.Background(), req)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestUserUseCase_GetUserByID(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		expectedUser := &entity.User{ID: "test123", Name: "Test User"}

		deps.Repo.On("FindByID", mock.Anything, "test123").Return(expectedUser, nil)

		result, err := uc.GetUserByID(context.Background(), "test123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test123", result.ID)

		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("FindByID", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))

		result, err := uc.GetUserByID(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrNotFound, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		_, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"

		result, err := uc.GetUserByID(context.Background(), sqlInjectionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrBadRequest, err)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		dbError := errors.New("database connection failed")

		deps.Repo.On("FindByID", mock.Anything, "db-error").Return(nil, dbError)

		result, err := uc.GetUserByID(context.Background(), "db-error")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, dbError, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestUserUseCase_GetAllUsers(t *testing.T) {
	t.Run("Success - With Users", func(t *testing.T) {
		deps, uc := setupUserTest()
		mockUsers := []*entity.User{
			{ID: "user1", Name: "User One"},
			{ID: "user2", Name: "User Two"},
		}
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		deps.Repo.On("FindAll", mock.Anything, req).Return(mockUsers, nil)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)

		deps.Repo.AssertExpectations(t)
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		deps, uc := setupUserTest()
		req := &model.GetUserListRequest{Page: 1, Limit: 10}
		deps.Repo.On("FindAll", mock.Anything, req).Return([]*entity.User{}, nil)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, result)

		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		dbError := errors.New("database connection failed")
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		deps.Repo.On("FindAll", mock.Anything, req).Return(nil, dbError)

		result, err := uc.GetAllUsers(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, exception.ErrInternalServer, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestUserUseCase_Current(t *testing.T) {
	t.Run("Success - User Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		expectedUser := &entity.User{ID: "current-user", Name: "Current User"}
		testReq := &model.GetUserRequest{ID: "current-user"}

		deps.Repo.On("FindByID", mock.Anything, "current-user").Return(expectedUser, nil)

		result, err := uc.Current(context.Background(), testReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "current-user", result.ID)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		testReq := &model.GetUserRequest{ID: "nonexistent"}

		deps.Repo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		result, err := uc.Current(context.Background(), testReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		deps.Repo.AssertExpectations(t)
	})
}

func TestUserUseCase_Update(t *testing.T) {
	t.Run("Success - User Updated", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Name: "Updated User",
		}

		existingUser := &entity.User{
			ID:   "user123",
			Name: "Original User",
		}

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == "user123" && u.Name == "Updated User"
		})).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		result, err := uc.Update(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "user123", result.ID)
		assert.Equal(t, "Updated User", result.Name)

		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
	})

	t.Run("Success - Update Username", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Username: "newuser",
		}
		existingUser := &entity.User{ID: "user123", Username: "olduser"}

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.Repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		_, err := uc.Update(context.Background(), request)
		assert.NoError(t, err)
	})

	t.Run("Error - Username Conflict", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Username: "exists",
		}
		existingUser := &entity.User{ID: "user123", Username: "olduser"}

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.Repo.On("FindByUsername", mock.Anything, "exists").Return(&entity.User{Username: "exists"}, nil)

		_, err := uc.Update(context.Background(), request)
		assert.ErrorIs(t, err, exception.ErrConflict)
	})

	t.Run("Update Password - Success", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Password: "newpassword123",
		}

		existingUser := &entity.User{ID: "user123"}

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		_, err := uc.Update(context.Background(), request)
		assert.NoError(t, err)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		updateReq := &model.UpdateUserRequest{
			ID:   "nonexistent",
			Name: "New Name",
		}

		deps.Repo.On("FindByID", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

		result, err := uc.Update(context.Background(), updateReq)

		assert.ErrorIs(t, err, exception.ErrNotFound)
		assert.Nil(t, result)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Update Fails", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{ID: "user123", Name: "Updated"}
		existingUser := &entity.User{ID: "user123"}
		dbErr := errors.New("db update failed")

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(dbErr)

		_, err := uc.Update(context.Background(), request)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	actorUserID := "admin-user-id"
	cleanID := "019b9150-304e-79d0-aa16-4a2b44347a08"
	deleteReq := &model.DeleteUserRequest{ID: cleanID}

	t.Run("Success - User Deleted", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID, Username: "deletedUser"}, nil)
		deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(nil, errors.New("user not found"))

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrNotFound, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - SQL Injection Attempt", func(t *testing.T) {
		_, uc := setupUserTest()
		sqlInjectionID := "1'; DROP TABLE users;--"
		deleteReqSqli := &model.DeleteUserRequest{ID: sqlInjectionID}

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReqSqli)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrBadRequest, err)
	})

	t.Run("Error - Database Error During Delete", func(t *testing.T) {
		deps, uc := setupUserTest()
		dbError := errors.New("internal server error")

		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID}, nil)
		deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(dbError)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})
}

func TestUserUseCase_GetAllUsersDynamic(t *testing.T) {
	t.Run("Success - With Dynamic Filter", func(t *testing.T) {
		deps, uc := setupUserTest()
		mockUsers := []*entity.User{
			{ID: "user1", Name: "Dynamic User 1"},
			{ID: "user2", Name: "Dynamic User 2"},
		}

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Dynamic"},
			},
		}

		deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(mockUsers, nil)

		result, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "Dynamic User 1", result[0].Name)

		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		dbError := errors.New("database error")
		expectedError := exception.ErrInternalServer

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Error"},
			},
		}

		deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(nil, dbError)

		result, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		deps.Repo.AssertExpectations(t)
	})
}

func TestUserUseCase_UpdateStatus(t *testing.T) {
	t.Run("Success - Active", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		status := entity.UserStatusActive

		deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "UPDATE_STATUS"
		})).Return(nil)

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
	})

	t.Run("Success - Banned (Revoke Sessions Failure)", func(t *testing.T) {
    deps, uc := setupUserTest()
    userID := "user123"
    status := entity.UserStatusBanned
    deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
    deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
    deps.AuthUC.On("RevokeAllSessions", mock.Anything, userID).Return(errors.New("revoke failed"))
    deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)
    err := uc.UpdateStatus(context.Background(), userID, status)
    assert.NoError(t, err)
    deps.Repo.AssertExpectations(t)
    deps.AuthUC.AssertExpectations(t)
})
t.Run("Error - Invalid Status", func(t *testing.T) {
    _, uc := setupUserTest()
    err := uc.UpdateStatus(context.Background(), "user123", "invalid_status")
    assert.Equal(t, exception.ErrValidationError, err)
})
t.Run("Error - User Not Found", func(t *testing.T) {
    deps, uc := setupUserTest()
    deps.Repo.On("FindByID", mock.Anything, "unknown").Return(nil, errors.New("user not found"))
    err := uc.UpdateStatus(context.Background(), "unknown", entity.UserStatusActive)
    assert.Equal(t, exception.ErrNotFound, err)
})
// DARI guardian-coverage-boost:
t.Run("Error - Update Status Fails", func(t *testing.T) {
    deps, uc := setupUserTest()
    userID := "user123"
    status := entity.UserStatusActive
    dbErr := errors.New("db error")
    deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
    deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(dbErr)
    err := uc.UpdateStatus(context.Background(), userID, status)
    assert.Equal(t, exception.ErrInternalServer, err)
})
// DARI dev:
t.Run("Audit Log Error", func(t *testing.T) {
    deps, uc := setupUserTest()
    userID := "user123"
    status := entity.UserStatusActive
    deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
    deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
    deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))
    err := uc.UpdateStatus(context.Background(), userID, status)
    assert.NoError(t, err)
})
}
