package test

import (
	"context"
	"errors"
	"io"
	"testing"

	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	txMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type userTestDeps struct {
	Repo     *userMocks.MockUserRepository
	TM       *txMocks.MockWithTransactionManager
	Enforcer *permissionMocks.IEnforcer
	AuditUC  *auditMocks.MockAuditUseCase
	AuthUC   *authMocks.MockAuthUseCase
}

func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
	deps := &userTestDeps{
		Repo:     new(userMocks.MockUserRepository),
		TM:       new(txMocks.MockWithTransactionManager),
		Enforcer: new(permissionMocks.IEnforcer),
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

	req := &model.RegisterUserRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	}

	deps.Repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found"))
	deps.Repo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == "newuser" && u.Email == "new@example.com"
	})).Return(nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything, "role:user").Return(true, nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	res, err := uc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "newuser", res.Username)
	assert.Equal(t, "new@example.com", res.Email)

	deps.Repo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestUserUseCase_Create_Conflict(t *testing.T) {
	deps, uc := setupUserTest()

	req := &model.RegisterUserRequest{Username: "existing", Email: "e@e.com", Password: "p"}

	deps.Repo.On("FindByUsername", mock.Anything, "existing").Return(&entity.User{ID: "1"}, nil)

	res, err := uc.Create(context.Background(), req)

	assert.ErrorIs(t, err, exception.ErrConflict)
	assert.Nil(t, res)
}

func TestUserUseCase_GetUserByID_Success(t *testing.T) {
	deps, uc := setupUserTest()
	id := "user-123"
	user := &entity.User{ID: id, Username: "test"}

	deps.Repo.On("FindByID", mock.Anything, id).Return(user, nil)

	res, err := uc.GetUserByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, id, res.ID)
}

func TestUserUseCase_GetUserByID_NotFound(t *testing.T) {
	deps, uc := setupUserTest()
	id := "unknown"

	deps.Repo.On("FindByID", mock.Anything, id).Return(nil, errors.New("user not found"))

	res, err := uc.GetUserByID(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrNotFound)
	assert.Nil(t, res)
}

func TestUserUseCase_GetUserByID_SQLInjection(t *testing.T) {
	_, uc := setupUserTest()
	id := "' OR 1=1 --"

	res, err := uc.GetUserByID(context.Background(), id)

	assert.ErrorIs(t, err, exception.ErrBadRequest)
	assert.Nil(t, res)
}

func TestUserUseCase_UpdateStatus_Success(t *testing.T) {
	deps, uc := setupUserTest()
	id := "u1"
	status := entity.UserStatusBanned

	deps.Repo.On("FindByID", mock.Anything, id).Return(&entity.User{ID: id}, nil)
	deps.Repo.On("UpdateStatus", mock.Anything, id, status).Return(nil)
	deps.AuthUC.On("RevokeAllSessions", mock.Anything, id).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := uc.UpdateStatus(context.Background(), id, status)

	assert.NoError(t, err)
	deps.AuthUC.AssertCalled(t, "RevokeAllSessions", mock.Anything, id)
}

func TestUserUseCase_UpdateStatus_InvalidStatus(t *testing.T) {
	_, uc := setupUserTest()
	err := uc.UpdateStatus(context.Background(), "u1", "INVALID")
	assert.ErrorIs(t, err, exception.ErrValidationError)
}

func TestUserUseCase_DeleteUser_Success(t *testing.T) {
	deps, uc := setupUserTest()
	id := "u1"

	deps.Repo.On("FindByID", mock.Anything, id).Return(&entity.User{ID: id, Username: "del"}, nil)
	deps.Repo.On("Delete", mock.Anything, id).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	req := &model.DeleteUserRequest{ID: id}
	err := uc.DeleteUser(context.Background(), "admin", req)

	assert.NoError(t, err)
}

func TestUserUseCase_Update_Success(t *testing.T) {
	deps, uc := setupUserTest()
	id := "u1"
	req := &model.UpdateUserRequest{ID: id, Name: "New Name", Username: "newname"}

	existing := &entity.User{ID: id, Username: "oldname", Name: "Old Name"}

	deps.Repo.On("FindByID", mock.Anything, id).Return(existing, nil)
	deps.Repo.On("FindByUsername", mock.Anything, "newname").Return(nil, errors.New("not found"))
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Name == "New Name" && u.Username == "newname"
	})).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	res, err := uc.Update(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", res.Name)
}

func TestUserUseCase_Update_XSSSanitization(t *testing.T) {
	deps, uc := setupUserTest()
	id := "u1"
	req := &model.UpdateUserRequest{ID: id, Name: "<script>alert(1)</script>User"}

	existing := &entity.User{ID: id, Username: "u", Name: "O"}

	deps.Repo.On("FindByID", mock.Anything, id).Return(existing, nil)
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		// SanitizeString uses html.EscapeString, so we expect escaped HTML
		return u.Name == "&lt;script&gt;alert(1)&lt;/script&gt;User"
	})).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	res, err := uc.Update(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "&lt;script&gt;alert(1)&lt;/script&gt;User", res.Name)
}
