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
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
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
	Storage  *storageMocks.MockProvider
}

func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
	deps := &userTestDeps{
		Repo:     new(mocks.MockUserRepository),
		TM:       new(mocking.MockWithTransactionManager),
		Enforcer: new(permMocks.IEnforcer),
		AuditUC:  new(auditMocks.MockAuditUseCase),
		AuthUC:   new(authMocks.MockAuthUseCase),
		Storage:  new(storageMocks.MockProvider),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.FatalLevel)

	uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC, deps.AuthUC, deps.Storage)

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
		deps.Storage.On("UploadFile", mock.Anything, "avatars/user123.png", "image/png").Return("url", nil)
		deps.Repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))
		deps.Storage.On("DeleteFile", mock.Anything, "avatars/user123.png").Return(nil)

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

	deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, "global").Return(true, nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	result, err := uc.Create(context.Background(), testReq)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedName, result.Name)
	deps.Repo.AssertExpectations(t)
}

func TestUserUseCase_Create_PasswordTooLong(t *testing.T) {
	deps, uc := setupUserTest()
	// 73 chars
	longPassword := "1234567890123456789012345678901234567890123456789012345678901234567890123"
	req := &model.RegisterUserRequest{
		Username: "user", Email: "test@example.com", Password: longPassword, Name: "Test",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "user").Return(nil, gorm.ErrRecordNotFound)
	deps.Repo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, gorm.ErrRecordNotFound)

	_, err := uc.Create(context.Background(), req)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestUserUseCase_Update_PasswordTooLong(t *testing.T) {
	deps, uc := setupUserTest()
	longPassword := "1234567890123456789012345678901234567890123456789012345678901234567890123"
	request := &model.UpdateUserRequest{
		ID: "user123", Password: longPassword,
	}

	deps.Repo.On("FindByID", mock.Anything, "user123").Return(&entity.User{ID: "user123"}, nil)

	_, err := uc.Update(context.Background(), request)
	assert.Equal(t, exception.ErrInternalServer, err)
}
