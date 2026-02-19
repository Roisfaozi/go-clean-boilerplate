package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResetPassword_Failure_RevokeSessions(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	token := "valid-token"
	resetToken := &authEntity.PasswordResetToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)

	// Transaction succeeds (User updated, Token deleted)
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)

	// Expect Audit Log for Password Reset Success
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "PASSWORD_RESET_SUCCESS"
	})).Return(nil)

	// Expect Audit Log for Revoke All Sessions (called inside RevokeAllSessions)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "REVOKE_ALL_SESSIONS"
	})).Return(nil)

	// RevokeAllSessions FAILS
	revokeErr := errors.New("redis down")
	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(revokeErr)

	// We expect an error to be returned, which triggers transaction rollback in the real implementation.
	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to revoke sessions")

	// Verify that Update was called (meaning password was changed in DB)
	deps.userRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*entity.User"))
	// Verify token was deleted
	deps.tokenRepo.AssertCalled(t, "DeleteByEmail", mock.Anything, user.Email)
}
