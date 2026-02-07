package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUseCase_DeleteUser_RoleBackupFailure_ShouldFail(t *testing.T) {
	deps, uc := setupUserTest()
	deleteReq := &model.DeleteUserRequest{ID: "fail-backup-id"}
	actorUserID := "admin-user"

	// Mock User Found
	deps.Repo.On("FindByID", mock.Anything, deleteReq.ID).Return(&entity.User{ID: deleteReq.ID}, nil)

	// Mock Transaction
	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	// Mock Delete Success
	deps.Repo.On("Delete", mock.Anything, deleteReq.ID).Return(nil)

	// Mock Role Backup Failure
	deps.Enforcer.On("GetRolesForUser", deleteReq.ID, "global").Return([]string{}, errors.New("casbin error"))


	// Currently the code swallows this error and logs a warning.
	// We want it to fail closed to prevent role loss.
	// So we expect an error here.

	// Note: Since the current implementation DOES NOT fail, this test will FAIL initially.
	// This is intentional to prove the bug/change requirement.
	err := uc.DeleteUser(context.Background(), actorUserID, deleteReq)

	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
	deps.Enforcer.AssertExpectations(t)
}
