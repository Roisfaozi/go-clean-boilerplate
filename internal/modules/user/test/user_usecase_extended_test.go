package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	permMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	userUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	storageMocks "github.com/Roisfaozi/go-clean-boilerplate/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserExtendedTest() (*userTestDeps, userUseCase.UserUseCase) {
	mockEnforcer := new(permMocks.IEnforcer)
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

	var enf permissionUseCase.IEnforcer = deps.Enforcer
	uc := userUseCase.NewUserUseCase(deps.TM, log, deps.Repo, enf, deps.AuditUC, deps.AuthUC, deps.Storage)

	return deps, uc
}

func TestUserUseCase_Update_Extended(t *testing.T) {
	deps, uc := setupUserExtendedTest()
	req := &model.UpdateUserRequest{ID: "user1", Name: "New Name"}
	user := &entity.User{ID: "user1"}

	deps.Repo.On("FindByID", mock.Anything, "user1").Return(user, nil)
	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(exception.ErrInternalServer)

	deps.Repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	_, err := uc.Update(context.Background(), req)
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestUserUseCase_UpdateStatus_Extended(t *testing.T) {
	deps, uc := setupUserExtendedTest()

	t.Run("Invalid_Status", func(t *testing.T) {
		err := uc.UpdateStatus(context.Background(), "user1", "invalid")
		assert.ErrorIs(t, err, exception.ErrValidationError)
	})

	t.Run("Revoke_Error_Fails_Tx", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "user1").Return(&entity.User{ID: "user1"}, nil)
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(exception.ErrInternalServer)

		deps.Repo.On("UpdateStatus", mock.Anything, "user1", entity.UserStatusBanned).Return(nil)
		deps.AuthUC.On("RevokeAllSessions", mock.Anything, "user1").Return(errors.New("revoke error"))

		err := uc.UpdateStatus(context.Background(), "user1", entity.UserStatusBanned)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestUserUseCase_UpdateAvatar_Extended(t *testing.T) {
	deps, uc := setupUserExtendedTest()
	deps.Repo.On("FindByID", mock.Anything, "user1").Return(&entity.User{ID: "user1"}, nil)

	t.Run("File_Read_Error", func(t *testing.T) {
		// Pass a reader that fails
		badReader := &errReader{}
		_, err := uc.UpdateAvatar(context.Background(), "user1", badReader, "f.png", "image/png")
		assert.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("Invalid_Content_Type", func(t *testing.T) {
		// Mock reader with text content
		textReader := strings.NewReader("not an image")
		_, err := uc.UpdateAvatar(context.Background(), "user1", textReader, "f.txt", "text/plain")
		assert.ErrorIs(t, err, exception.ErrValidationError)
	})
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestUserUseCase_DeleteUser_Extended(t *testing.T) {
	deps, uc := setupUserExtendedTest()
	userID := "user1"
	user := &entity.User{ID: userID, Username: "u1"}

	t.Run("Role_Restore_Failure_On_Rollback", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, userID).Return(user, nil)

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(exception.ErrInternalServer)

		deps.Repo.On("Delete", mock.Anything, userID).Return(nil)

		// Enforcer Mocks
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("GetRolesForUser", userID, "global").Return([]string{"role:user"}, nil)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", "global").Return(true, nil)

		// Audit Failure triggers rollback
		deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit fail"))

		// Compensation Failure
		deps.Enforcer.On("AddGroupingPolicy", userID, "role:user", "global").Return(false, errors.New("restore fail"))

		err := uc.DeleteUser(context.Background(), "admin", &model.DeleteUserRequest{ID: userID})
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})
}
