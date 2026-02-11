package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"io"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	mock_permission "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
)

// setupVerificationTest creates test dependencies for verification tests
func setupVerificationTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)

	deps := &testDependencies{
		jwtManager:      jwtManager,
		tokenRepo:       new(mock_auth.MockTokenRepository),
		userRepo:        new(mock_user.MockUserRepository),
		orgRepo:         new(mock_org.MockOrganizationRepository),
		tm:              new(mocking.MockWithTransactionManager),
		wsManager:       new(mocking.MockManager),
		enforcer:        new(mock_permission.IEnforcer),
		log:             logrus.New(),
		auditUC:         new(auditMocks.MockAuditUseCase),
		taskDistributor: new(mocking.MockTaskDistributor),
	}

	deps.log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		5,              // MaxLoginAttempts
		30*time.Minute, // LockoutDuration
		deps.jwtManager,
		deps.tokenRepo,
		deps.userRepo,
		deps.orgRepo,
		deps.tm,
		deps.log,
		deps.wsManager,
		nil, // sseManager
		deps.enforcer,
		deps.auditUC,
		deps.taskDistributor,
	)

	return authService, deps
}

// ============================================================================
// REQUEST VERIFICATION TESTS
// ============================================================================

// ✅ POSITIVE CASE
func TestAuthUseCase_RequestVerification_Success(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	userID := "user-123"
	user := &entity.User{
		ID:              userID,
		Username:        "testuser",
		Email:           "test@example.com",
		EmailVerifiedAt: nil, // Not verified yet
	}

	// Mock FindByID
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock SaveVerificationToken
	deps.tokenRepo.On("SaveVerificationToken", ctx, mock.MatchedBy(func(token *authEntity.EmailVerificationToken) bool {
		return token.Email == user.Email && len(token.Token) == 32 // 16 bytes = 32 hex chars
	})).Return(nil)

	// Mock Task Distributor
	deps.taskDistributor.On("DistributeTaskSendEmail", ctx, mock.MatchedBy(func(payload *tasks.SendEmailPayload) bool {
		return payload.To == user.Email && payload.Subject == "Verify Your Email Address"
	})).Return(nil)

	// Mock Audit Log
	deps.auditUC.On("LogActivity", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID &&
			req.Action == "VERIFICATION_EMAIL_REQUESTED" &&
			req.Entity == "User"
	})).Return(nil)

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert
	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

// ❌ NEGATIVE CASES
func TestAuthUseCase_RequestVerification_UserNotFound(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	userID := "nonexistent-user"

	// Mock FindByID - User not found
	deps.userRepo.On("FindByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertNotCalled(t, "SaveVerificationToken")
}

func TestAuthUseCase_RequestVerification_AlreadyVerified(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	userID := "user-456"
	now := time.Now().UnixMilli()
	user := &entity.User{
		ID:              userID,
		Username:        "verifieduser",
		Email:           "verified@example.com",
		EmailVerifiedAt: &now, // Already verified
	}

	// Mock FindByID
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrAlreadyVerified, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertNotCalled(t, "SaveVerificationToken")
}

func TestAuthUseCase_RequestVerification_TokenGenerationError(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	userID := "user-789"
	user := &entity.User{
		ID:              userID,
		Username:        "testuser2",
		Email:           "test2@example.com",
		EmailVerifiedAt: nil,
	}

	// Mock FindByID
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock SaveVerificationToken - Error
	deps.tokenRepo.On("SaveVerificationToken", ctx, mock.Anything).
		Return(errors.New("database error"))

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskSendEmail")
}

func TestAuthUseCase_RequestVerification_TaskDistributorError(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	userID := "user-101"
	user := &entity.User{
		ID:              userID,
		Username:        "testuser3",
		Email:           "test3@example.com",
		EmailVerifiedAt: nil,
	}

	// Mock FindByID
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock SaveVerificationToken
	deps.tokenRepo.On("SaveVerificationToken", ctx, mock.Anything).Return(nil)

	// Mock Task Distributor - Error (should not fail the request)
	deps.taskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).
		Return(errors.New("queue is full"))

	// Mock Audit Log
	deps.auditUC.On("LogActivity", ctx, mock.Anything).Return(nil)

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert - Should still succeed even if email task fails
	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

// 🔄 EDGE CASE
func TestAuthUseCase_RequestVerification_EmptyUserID(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	userID := ""

	// Mock FindByID - Should fail with empty ID
	deps.userRepo.On("FindByID", ctx, userID).Return(nil, errors.New("invalid user id"))

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert
	assert.Error(t, err)
	deps.userRepo.AssertExpectations(t)
}

// ============================================================================
// VERIFY EMAIL TESTS
// ============================================================================

// ✅ POSITIVE CASE
func TestAuthUseCase_VerifyEmail_Success(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	token := "valid-verification-token-32chars"
	email := "test@example.com"
	userID := "user-202"

	verificationToken := &authEntity.EmailVerificationToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour).UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
	}

	user := &entity.User{
		ID:              userID,
		Username:        "testuser",
		Email:           email,
		EmailVerifiedAt: nil, // Not verified yet
	}

	// Mock FindVerificationToken
	deps.tokenRepo.On("FindVerificationToken", ctx, token).Return(verificationToken, nil)

	// Mock FindByEmail
	deps.userRepo.On("FindByEmail", ctx, email).Return(user, nil)

	// Mock Transaction
	deps.tm.On("WithinTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

	// Mock Update
	deps.userRepo.On("Update", ctx, mock.MatchedBy(func(u *entity.User) bool {
		return u.ID == userID && u.EmailVerifiedAt != nil
	})).Return(nil)

	// Mock DeleteVerificationTokenByEmail
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", ctx, email).Return(nil)

	// Mock Audit Log
	deps.auditUC.On("LogActivity", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID &&
			req.Action == "EMAIL_VERIFIED" &&
			req.Entity == "User"
	})).Return(nil)

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.tm.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

// ❌ NEGATIVE CASES
func TestAuthUseCase_VerifyEmail_InvalidToken(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	token := "invalid-token"

	// Mock FindVerificationToken - Not found
	deps.tokenRepo.On("FindVerificationToken", ctx, token).
		Return(nil, errors.New("token not found"))

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidVerificationToken, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertNotCalled(t, "FindByEmail")
}

func TestAuthUseCase_VerifyEmail_ExpiredToken(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	token := "expired-token"
	email := "test@example.com"

	verificationToken := &authEntity.EmailVerificationToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(), // Expired
		CreatedAt: time.Now().Add(-25 * time.Hour).UnixMilli(),
	}

	// Mock FindVerificationToken
	deps.tokenRepo.On("FindVerificationToken", ctx, token).Return(verificationToken, nil)

	// Mock DeleteVerificationTokenByEmail (cleanup)
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", ctx, email).Return(nil)

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidVerificationToken, err)
	deps.tokenRepo.AssertExpectations(t)
}

func TestAuthUseCase_VerifyEmail_UserNotFound(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	token := "valid-token"
	email := "deleted@example.com"

	verificationToken := &authEntity.EmailVerificationToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour).UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
	}

	// Mock FindVerificationToken
	deps.tokenRepo.On("FindVerificationToken", ctx, token).Return(verificationToken, nil)

	// Mock FindByEmail - User deleted
	deps.userRepo.On("FindByEmail", ctx, email).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidVerificationToken, err)
	deps.userRepo.AssertExpectations(t)
}

func TestAuthUseCase_VerifyEmail_DatabaseUpdateError(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	token := "valid-token"
	email := "test@example.com"
	userID := "user-303"

	verificationToken := &authEntity.EmailVerificationToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour).UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
	}

	user := &entity.User{
		ID:              userID,
		Username:        "testuser",
		Email:           email,
		EmailVerifiedAt: nil,
	}

	// Mock FindVerificationToken
	deps.tokenRepo.On("FindVerificationToken", ctx, token).Return(verificationToken, nil)

	// Mock FindByEmail
	deps.userRepo.On("FindByEmail", ctx, email).Return(user, nil)

	// Mock Transaction - Failure
	dbErr := errors.New("database connection lost")
	deps.tm.On("WithinTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(dbErr)

	// Mock Update - Error
	deps.userRepo.On("Update", ctx, mock.Anything).Return(dbErr)

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	deps.tm.AssertExpectations(t)
	deps.auditUC.AssertNotCalled(t, "LogActivity")
}

// 🔄 EDGE CASE
func TestAuthUseCase_VerifyEmail_AlreadyVerified(t *testing.T) {
	authService, deps := setupVerificationTest(t)
	ctx := context.Background()

	token := "valid-token"
	email := "verified@example.com"
	userID := "user-404"
	now := time.Now().UnixMilli()

	verificationToken := &authEntity.EmailVerificationToken{
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour).UnixMilli(),
		CreatedAt: time.Now().UnixMilli(),
	}

	user := &entity.User{
		ID:              userID,
		Username:        "verifieduser",
		Email:           email,
		EmailVerifiedAt: &now, // Already verified
	}

	// Mock FindVerificationToken
	deps.tokenRepo.On("FindVerificationToken", ctx, token).Return(verificationToken, nil)

	// Mock FindByEmail
	deps.userRepo.On("FindByEmail", ctx, email).Return(user, nil)

	// Mock DeleteVerificationTokenByEmail (cleanup)
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", ctx, email).Return(nil)

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, usecase.ErrAlreadyVerified, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}
