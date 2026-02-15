package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
