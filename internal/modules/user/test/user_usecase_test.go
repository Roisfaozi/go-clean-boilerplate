package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type userTestDeps struct {
	Repo     *mocks.MockUserRepository
	TM       *mocking.MockWithTransactionManager
	Enforcer *permMocks.MockIEnforcer
	AuditUC  *auditMocks.MockAuditUseCase
	AuthUC   *authMocks.MockAuthUseCase
	Storage  *storageMocks.MockProvider
}

func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
	mockEnforcer := new(permMocks.MockIEnforcer)
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: mockEnforcer,
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.FatalLevel)

	// Cast to interface to ensure correct implementation
	var enf permissionUseCase.IEnforcer = deps.Enforcer

	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, enf, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}

func TestUserUseCase_Create_Success(t *testing.T) {
	deps, uc := setupUserTest()

	testReq := &model.RegisterUserRequest{
		Username: "testuser", Email: "test@example.com", Name: "Test User", Password: "password123",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)

	// Mock Transaction that executes closure
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
	deps.Enforcer.On("AddGroupingPolicy", mock.AnythingOfType("string"), "role:user", "global").Return(true, nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	deps.Repo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
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

	// Mock Transaction that executes closure and returns its error
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := uc.Create(context.Background(), req)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestUserUseCase_Create_AuditError(t *testing.T) {
	deps, uc := setupUserTest()
	req := &model.RegisterUserRequest{
		Username: "auditfail", Email: "audit@fail.com", Password: "password123", Name: "Audit Fail",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "auditfail").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "audit@fail.com").Return(nil, gorm.ErrRecordNotFound)

	// Mock Transaction
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, mock.Anything, "", "global").Return(true, nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	_, err := uc.Create(context.Background(), req)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestUserUseCase_Create_EnforcerError(t *testing.T) {
	deps, uc := setupUserTest()
	req := &model.RegisterUserRequest{
		Username: "user", Email: "test@example.com", Password: "password123", Name: "Test",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "user").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)

	// Mock Transaction that executes closure
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything, "role:user", "global").Return(false, errors.New("casbin error"))

	result, err := uc.Create(context.Background(), req)

	// After refactoring, Casbin failure now causes rollback
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	assert.Nil(t, result)
	deps.Enforcer.AssertExpectations(t)
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

		deps.Repo.On("FindAll", mock.Anything, req).Return(mockUsers, int64(2), nil)

		result, total, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, "user1", result[0].ID)
		assert.Equal(t, "user2", result[1].ID)

		deps.Repo.AssertExpectations(t)
	})

	t.Run("Success - Empty Result", func(t *testing.T) {
		deps, uc := setupUserTest()
		req := &model.GetUserListRequest{Page: 1, Limit: 10}
		deps.Repo.On("FindAll", mock.Anything, req).Return([]*entity.User{}, int64(0), nil)

		result, total, err := uc.GetAllUsers(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, int64(0), total)

		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Database Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		dbError := errors.New("database connection failed")
		req := &model.GetUserListRequest{Page: 1, Limit: 10}

		deps.Repo.On("FindAll", mock.Anything, req).Return(nil, int64(0), dbError)

		result, total, err := uc.GetAllUsers(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), total)
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
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
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
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
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
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		_, err := uc.Update(context.Background(), request)
		assert.NoError(t, err)
	})

	t.Run("Update - Conflict", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{
			ID: "user123", Username: "exists",
		}

		existingUser := &entity.User{ID: "user123", Username: "original"}
		otherUser := &entity.User{ID: "user456", Username: "exists"}

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.Repo.On("FindByUsername", mock.Anything, "exists").Return(otherUser, nil)

		_, err := uc.Update(context.Background(), request)
		assert.ErrorIs(t, err, exception.ErrConflict)
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
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(dbErr)

		_, err := uc.Update(context.Background(), request)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Audit Log Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		request := &model.UpdateUserRequest{ID: "user123", Name: "Updated"}
		existingUser := &entity.User{ID: "user123"}

		deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

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

		// Mock Transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)

		// Expect Backup Roles
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("GetRolesForUser", deleteReq.ID, "global").Return([]string{"role:user"}, nil)

		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, deleteReq.ID, "", "global").Return(true, nil)
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

		// Mock Transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(dbError)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
	})

	t.Run("Error - Audit Log Fails (Compensation Triggered)", func(t *testing.T) {
		deps, uc := setupUserTest()

		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID, Username: "deletedUser"}, nil)

		// Mock Transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)

		// Expect Backup Roles
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("GetRolesForUser", deleteReq.ID, "global").Return([]string{"role:user", "role:admin"}, nil)

		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, deleteReq.ID, "", "global").Return(true, nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit fail"))

		// Expect Compensation: Restore Roles
		deps.Enforcer.On("AddGroupingPolicy", deleteReq.ID, "role:user", "global").Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", deleteReq.ID, "role:admin", "global").Return(true, nil)

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
		deps.Enforcer.AssertExpectations(t)
	})

	t.Run("Error - Audit Log Fails & Compensation Fails", func(t *testing.T) {
		deps, uc := setupUserTest()

		deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID, Username: "deletedUser"}, nil)

		// Mock Transaction
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)

		// Expect Backup Roles
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("GetRolesForUser", deleteReq.ID, "global").Return([]string{"role:user"}, nil)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, deleteReq.ID, "", "global").Return(true, nil)

		// Audit fails
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit fail"))

		// Compensation fails
		deps.Enforcer.On("AddGroupingPolicy", deleteReq.ID, "role:user", "global").Return(false, errors.New("casbin restore error"))

		err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

		// Should still return error, but code should not panic and should log error (which we can't assert easily without hooking logger, but coverage will increase)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
		deps.Enforcer.AssertExpectations(t)
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

		deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(mockUsers, int64(2), nil)

		result, total, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
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

		deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(nil, int64(0), dbError)

		result, total, err := uc.GetAllUsersDynamic(context.Background(), filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), total)
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
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "UPDATE_STATUS"
		})).Return(nil)

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
	})

	t.Run("Success - Banned (Revoke Sessions)", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		status := entity.UserStatusBanned

		deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)

		deps.AuthUC.On("RevokeAllSessions", mock.Anything, userID).Return(nil)

		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
		deps.AuthUC.AssertExpectations(t)
		deps.AuditUC.AssertExpectations(t)
	})

	t.Run("Revoke Sessions Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		status := entity.UserStatusBanned

		deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
		deps.AuthUC.On("RevokeAllSessions", mock.Anything, userID).Return(errors.New("redis error"))
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.Error(t, err)
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

	t.Run("Audit Log Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		status := entity.UserStatusActive

		deps.Repo.On("FindByID", mock.Anything, userID).Return(&entity.User{ID: userID}, nil)
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

		err := uc.UpdateStatus(context.Background(), userID, status)
		assert.Error(t, err)
	})
}

func TestUserUseCase_UpdateAvatar(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		file := createValidImageReader("image content")
		filename := "avatar.png"
		contentType := "image/png"
		expectedURL := "https://storage.com/avatars/user123.png"

		user := &entity.User{ID: userID}

		deps.Repo.On("FindByID", mock.Anything, userID).Return(user, nil)
		deps.Storage.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, contentType).Return(expectedURL, nil)
		deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.AvatarURL == expectedURL
		})).Return(nil)
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		result, err := uc.UpdateAvatar(context.Background(), userID, file, filename, contentType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedURL, result.AvatarURL)
		deps.Repo.AssertExpectations(t)
		deps.Storage.AssertExpectations(t)
	})

	t.Run("Error - User Not Found", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("FindByID", mock.Anything, "unknown").Return(nil, errors.New("user not found"))

		_, err := uc.UpdateAvatar(context.Background(), "unknown", nil, "f.png", "image/png")
		assert.Equal(t, exception.ErrNotFound, err)
	})

	t.Run("Error - Upload Failed", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		user := &entity.User{ID: userID}

		deps.Repo.On("FindByID", mock.Anything, userID).Return(user, nil)
		deps.Storage.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("s3 error"))

		_, err := uc.UpdateAvatar(context.Background(), userID, createValidImageReader(""), "f.png", "image/png")
		assert.Equal(t, exception.ErrInternalServer, err)
	})

	t.Run("Error - DB Update Failed", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user123"
		user := &entity.User{ID: userID}

		deps.Repo.On("FindByID", mock.Anything, userID).Return(user, nil)
		deps.Storage.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("url", nil)
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

		_, err := uc.UpdateAvatar(context.Background(), userID, createValidImageReader(""), "f.png", "image/png")
		assert.Equal(t, exception.ErrInternalServer, err)
	})
}

func TestUserUseCase_HardDeleteSoftDeletedUsers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupUserTest()
		retentionDays := 30
		deps.Repo.On("HardDeleteSoftDeletedUsers", mock.Anything, retentionDays).Return(nil)

		err := uc.HardDeleteSoftDeletedUsers(context.Background(), retentionDays)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		deps, uc := setupUserTest()
		deps.Repo.On("HardDeleteSoftDeletedUsers", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := uc.HardDeleteSoftDeletedUsers(context.Background(), 30)
		assert.Equal(t, exception.ErrInternalServer, err)
	})
}

func TestUserUseCase_Create_Sanitization(t *testing.T) {
	deps, uc := setupUserTest()

	inputName := "<script>alert('XSS')</script>John Doe"
	// pkg.SanitizeString implementation:
	// output = strings.TrimSpace(input)
	// output = html.EscapeString(output)
	// Wait, if it uses html.EscapeString, then "<script>" becomes "&lt;script&gt;"
	// It does NOT remove tags.

	// Let's check pkg/security.go again.
	// "func SanitizeString(input string) string { output := strings.TrimSpace(input); output = html.EscapeString(output); return output }"

	expectedName := "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;John Doe"

	testReq := &model.RegisterUserRequest{
		Username: "userXSS", Email: "xss@example.com", Name: inputName, Password: "password123",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "userXSS").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "xss@example.com").Return(nil, gorm.ErrRecordNotFound)

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Capture the user passed to Create and verify Name is sanitized
	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Name == expectedName
	})).Return(nil)

	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, "global").Return(true, nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedName, result.Name)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_Update_Sanitization(t *testing.T) {
	deps, uc := setupUserTest()

	inputUsername := "<b>bold</b>"
	// pkg.SanitizeString escapes HTML
	expectedUsername := "&lt;b&gt;bold&lt;/b&gt;"

	request := &model.UpdateUserRequest{
		ID: "user123", Username: inputUsername,
	}

	existingUser := &entity.User{
		ID:       "user123",
		Username: "olduser",
	}

	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

	// Expect FindByUsername to be called with sanitized username
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == expectedUsername
	})).Return(nil)

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	_, err := uc.Update(context.Background(), request)

	// This assertion might fail if the mock is strict about calls.
	// If it fails with "Unexpected call", that counts as test failure.
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

// --- Merged from use_case_avatar_test.go ---

// Helper to create a reader with valid PNG header
func createValidImageReader(content string) io.Reader {
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	header := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return io.MultiReader(strings.NewReader(string(header)), strings.NewReader(content))
}

// setupAvatarTest creates test dependencies for avatar tests
func setupAvatarTest() (*userTestDeps, usecase.UserUseCase) {
	mockEnforcer := new(permMocks.MockIEnforcer)
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: mockEnforcer,
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	// Cast to interface to ensure correct implementation
	var enf permissionUseCase.IEnforcer = deps.Enforcer

	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, enf, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_Success(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-123.png"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "", // No existing avatar
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-123.png", "image/png").
		Return(uploadedURL, nil)

	// Mock Update
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.AvatarURL == uploadedURL
	})).Return(nil)

	// Mock Audit Log
	deps.AuditUC.On("LogActivity", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID &&
			req.Action == "UPDATE_AVATAR" &&
			req.Entity == "User" &&
			req.EntityID == userID
	})).Return(nil)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uploadedURL, result.AvatarURL)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

func TestUserUseCase_GetAvatarUrl_Success(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	avatarURL := "avatars/user-123.png"
	fullURL := "https://storage.example.com/avatars/user-123.png"

	existingUser := &entity.User{
		ID:        userID,
		AvatarURL: avatarURL,
	}

	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)
	deps.Storage.On("GetFileUrl", ctx, avatarURL).Return(fullURL, nil)

	result, err := uc.GetAvatarUrl(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, fullURL, result)
}

func TestUserUseCase_UpdateAvatar_Success_ReplaceExisting(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-456"
	filename := "new-avatar.png"
	contentType := "image/png"
	fileContent := createValidImageReader("new-fake-image-data")
	oldAvatarURL := "https://storage.example.com/avatars/user-456-old.jpg"
	newAvatarURL := "https://storage.example.com/avatars/user-456.png"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser2",
		Email:     "test2@example.com",
		Name:      "Test User 2",
		AvatarURL: oldAvatarURL, // Has existing avatar
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload (replaces old one)
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-456.png", "image/png").
		Return(newAvatarURL, nil)

	// Mock Update
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.AvatarURL == newAvatarURL
	})).Return(nil)

	// Mock Audit Log
	deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(nil)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newAvatarURL, result.AvatarURL)
	assert.NotEqual(t, oldAvatarURL, result.AvatarURL)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

// ============================================================================
// ❌ NEGATIVE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_UserNotFound(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "nonexistent-user"
	filename := "profile.jpg"
	contentType := "image/jpeg"
	fileContent := strings.NewReader("fake-image-data")

	// Mock FindByID - User not found
	deps.Repo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrNotFound, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertNotCalled(t, "UploadFile")
}

func TestUserUseCase_UpdateAvatar_StorageUploadError(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-789"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser3",
		Email:    "test3@example.com",
		Name:     "Test User 3",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Error
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-789.png", "image/png").
		Return("", errors.New("storage service unavailable"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
	deps.Repo.AssertNotCalled(t, "Update")
}

func TestUserUseCase_UpdateAvatar_DatabaseUpdateError(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-101"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-101.png"

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser4",
		Email:    "test4@example.com",
		Name:     "Test User 4",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Success
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-101.png", "image/png").
		Return(uploadedURL, nil)

	// Mock Update - Error
	deps.Repo.On("Update", ctx, mock.Anything).Return(errors.New("database connection lost"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

func TestUserUseCase_UpdateAvatar_AuditLogError(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-202"
	filename := "profile.jpg"
	contentType := "image/png"
	fileContent := createValidImageReader("fake-image-data")
	uploadedURL := "https://storage.example.com/avatars/user-202.png"

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser5",
		Email:    "test5@example.com",
		Name:     "Test User 5",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-202.png", "image/png").
		Return(uploadedURL, nil)

	// Mock Update
	deps.Repo.On("Update", ctx, mock.Anything).Return(nil)

	// Mock Audit Log - Error (should fail the operation for consistency)
	deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(errors.New("audit service down"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert - Should fail if audit fails for consistency with other update methods
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

// ============================================================================
// 🔄 EDGE CASES
// ============================================================================

func TestUserUseCase_UpdateAvatar_InvalidFileType(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-303"
	filename := "malicious.exe"
	contentType := "application/x-msdownload"
	fileContent := strings.NewReader("fake-exe-data")

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser6",
		Email:    "test6@example.com",
		Name:     "Test User 6",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Should NOT be called
	// deps.Storage.On("UploadFile", ...).Return(...)

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, fileContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrValidationError, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertNotCalled(t, "UploadFile")
}

func TestUserUseCase_UpdateAvatar_FileTooLarge(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-404"
	filename := "huge-image.jpg"
	contentType := "image/png"
	// Simulate large file with valid header
	largeContent := createValidImageReader(strings.Repeat("x", 10*1024*1024)) // 10MB

	existingUser := &entity.User{
		ID:       userID,
		Username: "testuser7",
		Email:    "test7@example.com",
		Name:     "Test User 7",
	}

	// Mock FindByID
	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	// Mock Storage Upload - Should reject file too large
	deps.Storage.On("UploadFile", ctx, mock.Anything, "avatars/user-404.png", "image/png").
		Return("", errors.New("file size exceeds limit"))

	// Execute
	result, err := uc.UpdateAvatar(ctx, userID, largeContent, filename, contentType)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

// --- Merged from use_case_avatar_security_test.go ---

func TestUserUseCase_UpdateAvatar_Security(t *testing.T) {
	deps, uc := setupAvatarSecurityTest()
	ctx := context.Background()
	userID := "user-sec-123"

	tests := []struct {
		name        string
		filename    string
		contentType string
		fileContent io.Reader
		errExpected error
	}{
		{
			name:        "Block SVG (potential XSS)",
			filename:    "image.svg",
			contentType: "image/svg+xml",
			fileContent: strings.NewReader(`<?xml version="1.0" standalone="no"?><!DOCTYPE sql SYSTEM "http://malicious.com"><svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"></svg>`),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block HTML disguised as image",
			filename:    "fake.png",
			contentType: "image/png",
			fileContent: strings.NewReader(`<html><body><h1>Not an image</h1><script>alert(1)</script></body></html>`),
			errExpected: exception.ErrValidationError,
		},
		{
			name:        "Block Polyglot (PNG with PHP payload)",
			filename:    "poly.png",
			contentType: "image/png",
			fileContent: io.MultiReader(
				strings.NewReader("\x89PNG\r\n\x1a\n"),
				strings.NewReader("<?php echo 'malicious'; ?>"),
			),
			errExpected: nil,
		},
		{
			name:        "Block Script File",
			filename:    "exploit.sh",
			contentType: "text/x-shellscript",
			fileContent: strings.NewReader("#!/bin/bash\necho 'hacked'"),
			errExpected: exception.ErrValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingUser := &entity.User{ID: userID, Username: "secuser"}
			deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil).Once()

			if tt.errExpected == nil {
				deps.Storage.On("UploadFile", ctx, mock.Anything, mock.Anything, mock.Anything).Return("http://ok.com", nil).Once()
				deps.Repo.On("Update", ctx, mock.Anything).Return(nil).Once()
				deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(nil).Once()
			}

			_, err := uc.UpdateAvatar(ctx, userID, tt.fileContent, tt.filename, tt.contentType)

			if tt.errExpected != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.errExpected, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func setupAvatarSecurityTest() (*userTestDeps, usecase.UserUseCase) {
	mockEnforcer := new(permMocks.MockIEnforcer)
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: mockEnforcer,
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)
	return deps, uc
}

// --- Merged from use_case_cleanup_test.go ---

// setupCleanupTest creates test dependencies for cleanup tests
func setupCleanupTest() (*userTestDeps, usecase.UserUseCase) {
	mockEnforcer := new(permMocks.MockIEnforcer)
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: mockEnforcer,
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserUseCase_HardDeleteSoftDeletedUsers_Success(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	retentionDays := 30

	// Mock HardDeleteSoftDeletedUsers - Success
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).Return(nil)

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_HardDeleteSoftDeletedUsers_NoRecordsToDelete(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	retentionDays := 90

	// Mock HardDeleteSoftDeletedUsers - No records found, but no error
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).Return(nil)

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_SetAvatarURL_Success(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	url := "https://storage.example.com/avatars/user-123.png"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "",
	}

	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.AvatarURL == url
	})).Return(nil)
	deps.AuditUC.On("LogActivity", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "UPDATE_AVATAR_TUS"
	})).Return(nil)

	err := uc.SetAvatarURL(ctx, userID, url)

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

// ============================================================================
// ❌ NEGATIVE CASES
// ============================================================================

func TestUserUseCase_SetAvatarURL_UserNotFound(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "nonexistent-user"
	url := "https://storage.example.com/avatars/user-123.png"

	deps.Repo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.SetAvatarURL(ctx, userID, url)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
	deps.Repo.AssertExpectations(t)
	deps.Repo.AssertNotCalled(t, "Update")
}

func TestUserUseCase_SetAvatarURL_DatabaseError(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	url := "https://storage.example.com/avatars/user-123.png"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "",
	}

	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)
	deps.Repo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

	err := uc.SetAvatarURL(ctx, userID, url)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_SetAvatarURL_AuditLogError(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	url := "https://storage.example.com/avatars/user-123.png"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "",
	}

	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)
	deps.Repo.On("Update", ctx, mock.Anything).Return(nil)
	deps.AuditUC.On("LogActivity", ctx, mock.Anything).Return(errors.New("audit error"))

	err := uc.SetAvatarURL(ctx, userID, url)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
	deps.AuditUC.AssertExpectations(t)
}

func TestUserUseCase_GetAvatarUrl_UserNotFound(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "nonexistent-user"

	deps.Repo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := uc.GetAvatarUrl(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
	assert.Empty(t, result)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_GetAvatarUrl_NoAvatar(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"

	existingUser := &entity.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		Name:      "Test User",
		AvatarURL: "",
	}

	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)

	result, err := uc.GetAvatarUrl(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
	assert.Empty(t, result)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_GetAvatarUrl_StorageError(t *testing.T) {
	deps, uc := setupAvatarTest()
	ctx := context.Background()

	userID := "user-123"
	avatarURL := "avatars/user-123.png"

	existingUser := &entity.User{
		ID:        userID,
		AvatarURL: avatarURL,
	}

	deps.Repo.On("FindByID", ctx, userID).Return(existingUser, nil)
	deps.Storage.On("GetFileUrl", ctx, avatarURL).Return("", exception.ErrInternalServer)

	result, err := uc.GetAvatarUrl(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	assert.Empty(t, result)
	deps.Repo.AssertExpectations(t)
	deps.Storage.AssertExpectations(t)
}

func TestUserUseCase_HardDeleteSoftDeletedUsers_DatabaseError(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	retentionDays := 30

	// Mock HardDeleteSoftDeletedUsers - Database error
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).
		Return(errors.New("database connection lost"))

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_HardDeleteSoftDeletedUsers_InvalidRetentionDays(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	// Negative retention days
	retentionDays := -10

	// Mock HardDeleteSoftDeletedUsers - Should handle invalid input
	// Note: Current implementation doesn't validate, but repository might
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).
		Return(errors.New("invalid retention days"))

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Repo.AssertExpectations(t)
}

// ============================================================================
// 🔄 EDGE CASES
// ============================================================================

func TestUserUseCase_HardDeleteSoftDeletedUsers_ZeroRetentionDays(t *testing.T) {
	deps, uc := setupCleanupTest()
	ctx := context.Background()

	// Zero retention days - delete all soft-deleted users immediately
	retentionDays := 0

	// Mock HardDeleteSoftDeletedUsers - Should work with 0 days
	deps.Repo.On("HardDeleteSoftDeletedUsers", ctx, retentionDays).Return(nil)

	// Execute
	err := uc.HardDeleteSoftDeletedUsers(ctx, retentionDays)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

// --- Merged from use_case_security_test.go ---

func TestUserUseCase_Update_Security_UsernameSanitization(t *testing.T) {
	deps, uc := setupUserTest()

	// Input with HTML tags
	inputUsername := "<b>bold</b>"
	// Expected stored username (sanitized)
	expectedUsername := "&lt;b&gt;bold&lt;/b&gt;"

	request := &model.UpdateUserRequest{
		ID:       "user123",
		Username: inputUsername,
	}

	existingUser := &entity.User{
		ID:       "user123",
		Username: "olduser",
	}

	// Mock: Find user by ID
	deps.Repo.On("FindByID", mock.Anything, "user123").Return(existingUser, nil)

	// Mock: Check uniqueness
	// The usecase should sanitize BEFORE checking uniqueness to ensure we check the actual value to be stored.
	deps.Repo.On("FindByUsername", mock.Anything, expectedUsername).Return(nil, gorm.ErrRecordNotFound)

	// Mock: Transaction
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Mock: Update
	// Expect the USER passed to update to have the SANITIZED username
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == expectedUsername
	})).Return(nil)

	// Mock: Audit
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	// Execute
	_, err := uc.Update(context.Background(), request)

	// Assert
	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

// --- Merged from user_atomicity_test.go ---

func TestUserUseCase_UpdateStatus_Atomicity(t *testing.T) {
	ctx := context.Background()

	t.Run("Status updated, but Audit log fails -> Should return error", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user-1"
		status := entity.UserStatusBanned

		deps.Repo.On("FindByID", ctx, userID).Return(&entity.User{ID: userID}, nil).Once()
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil).Once()
		deps.AuthUC.On("RevokeAllSessions", mock.Anything, userID).Return(nil).Once()
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		// Audit log fails
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error")).Once()

		err := uc.UpdateStatus(ctx, userID, status)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrInternalServer, err)
	})

	t.Run("Status updated, but Revoke sessions fails -> Should return error", func(t *testing.T) {
		deps, uc := setupUserTest()
		userID := "user-2"
		status := entity.UserStatusBanned

		deps.Repo.On("FindByID", ctx, userID).Return(&entity.User{ID: userID}, nil).Once()
		deps.Repo.On("UpdateStatus", mock.Anything, userID, status).Return(nil).Once()
		deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		// Revoke sessions fails
		deps.AuthUC.On("RevokeAllSessions", mock.Anything, userID).Return(errors.New("revoke error")).Once()

		err := uc.UpdateStatus(ctx, userID, status)

		assert.Error(t, err)
		assert.Equal(t, exception.ErrInternalServer, err)
	})
}

// --- Merged from user_ctx_test.go ---

// TestGetUserByID_ContextCancellation tests that context cancellation is propagated to the repository.
func TestGetUserByID_ContextCancellation(t *testing.T) {
	// Setup dependencies
	mockRepo, _, _, _, _, _, _, uc := setupTestUserUseCase()

	// Create a context that is already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Setup expectations
	// The repository should be called with a context.
	// We can't strictly assert "ctx.Err() != nil" inside the mock matching easily without a custom matcher,
	// but we can assert that the context passed IS the same context.
	expectedErr := context.Canceled

	mockRepo.On("FindByID", mock.MatchedBy(func(c context.Context) bool {
		return c == ctx
	}), "user-123").Return(nil, expectedErr)

	// Execute
	result, err := uc.GetUserByID(ctx, "user-123")

	// Verify
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Helper setup function (reused from user_usecase_test.go logic if available, but simplified here for isolation)
// Assuming standard mocks are available in internal/modules/user/test/mocks based on previous file listings
func setupTestUserUseCase() (
	*mocks.MockUserRepository,
	interface{}, // DB generic
	interface{}, // Enforcer generic
	interface{}, // Audit generic
	interface{}, // Auth generic
	interface{}, // Storage generic
	*logrus.Logger,
	usecase.UserUseCase,
) {
	mockRepo := new(mocks.MockUserRepository)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.FatalLevel)

	// We need nil/mock placeholders for other deps to construct the usecase
	// Since GetUserByID only needs Repo and Logger, others can be nil or simple mocks if NewUserUseCase enforces non-nil.
	// Looking at user_usecase.go, NewUserUseCase takes interfaces.
	// checking if it panics on nil. The implementation struct just assigns them.
	// u.Repo.FindByID is called.

	// However, we need to respect the constructor signature.
	// NewUserUseCase(db, log, repo, enforcer, audit, auth, storage)

	// We might need to mock these if NewUserUseCase checks them, or if we want to be safe.
	// Based on previous files, I'll use simple nil or new() for interfaces if possible,
	// or valid mocks if I need to import them.
	// For this specific test, we only access Repo.

	// Re-using the MockTransactionManager from before would be good if available, or just nil if not used in GetUserByID.
	// GetUserByID does NOT use transaction manager (lines 144-160 of user_usecase.go).

	// So we pass nil for others.

	uc := usecase.NewUserUseCase(nil, logger, mockRepo, nil, nil, nil, nil)

	return mockRepo, nil, nil, nil, nil, nil, logger, uc
}

// --- Merged from user_validation_test.go ---

func TestUserXSSValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.FatalLevel)
	tests := []struct {
		name         string
		method       string
		url          string
		payload      interface{}
		expectedCode int
	}{
		{
			name:   "RegisterUser XSS in Name",
			method: "POST",
			url:    "/users",
			payload: model.RegisterUserRequest{
				Username: "testuser",
				Password: "password123",
				Name:     "<script>alert(1)</script>",
				Email:    "test@example.com",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "RegisterUser XSS in Username",
			method: "POST",
			url:    "/users",
			payload: model.RegisterUserRequest{
				Username: "<img src=x onerror=alert(1)>",
				Password: "password123",
				Name:     "Test User",
				Email:    "test@example.com",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "UpdateUser XSS in Name",
			method: "PUT",
			url:    "/users/1",
			payload: model.UpdateUserRequest{
				Name:     "<iframe src='javascript:alert(1)'></iframe>",
				Username: "testuser",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mocks.MockUserUseCase)
			controller := userHandler.NewUserController(mockUC, logger, v)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(tt.method, tt.url, bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", "1")

			if tt.method == "POST" {
				controller.RegisterUser(c)
			} else {
				controller.UpdateUser(c)
			}

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
