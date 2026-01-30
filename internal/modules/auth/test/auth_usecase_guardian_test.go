package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	mock_permission "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Define specific struct for Guardian tests to be self-contained
type authGuardianTestDeps struct {
	jwtManager      *jwt.JWTManager
	tokenRepo       *mock_auth.MockTokenRepository
	userRepo        *mock_user.MockUserRepository
	tm              *mocking.MockWithTransactionManager
	wsManager       *mocking.MockManager
	enforcer        *mock_permission.IEnforcer
	log             *logrus.Logger
	auditUC         *auditMocks.MockAuditUseCase
	taskDistributor *mocking.MockTaskDistributor
}

func setupAuthGuardianTest(t *testing.T) (usecase.AuthUseCase, *authGuardianTestDeps) {
	jwtManager := jwt.NewJWTManager("secret", "refresh-secret", 15*time.Minute, 24*time.Hour)

	deps := &authGuardianTestDeps{
		jwtManager:      jwtManager,
		tokenRepo:       new(mock_auth.MockTokenRepository),
		userRepo:        new(mock_user.MockUserRepository),
		tm:              new(mocking.MockWithTransactionManager),
		wsManager:       new(mocking.MockManager),
		enforcer:        new(mock_permission.IEnforcer),
		log:             logrus.New(),
		auditUC:         new(auditMocks.MockAuditUseCase),
		taskDistributor: new(mocking.MockTaskDistributor),
	}

	deps.log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		deps.jwtManager,
		deps.tokenRepo,
		deps.userRepo,
		deps.tm,
		deps.log,
		deps.wsManager,
		nil,
		deps.enforcer,
		deps.auditUC,
		deps.taskDistributor,
	)

	return authService, deps
}

func createGuardianTestUser(password string) (*userEntity.User, string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &userEntity.User{
		ID:       "user-test-id",
		Username: "testuser",
		Name:     "Test User",
		Password: string(hashedPassword),
		Email:    "test@example.com",
		Status:   userEntity.UserStatusActive,
	}, password
}

// TestAuthUseCase_Edge_UnicodeInUsername tests handling of Unicode characters in username.
func TestAuthUseCase_Edge_UnicodeInUsername(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	unicodeUsername := "ユーザー名"
	user, password := createGuardianTestUser("password123")
	user.Username = unicodeUsername
	loginReq := model.LoginRequest{Username: unicodeUsername, Password: password}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, unicodeUsername).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, unicodeUsername).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, unicodeUsername).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{"role:user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.wsManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return()

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	})).Return(nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Equal(t, unicodeUsername, loginResp.User.Username)
	deps.userRepo.AssertExpectations(t)
}

// TestAuthUseCase_Edge_LongUsername tests handling of extremely long usernames.
func TestAuthUseCase_Edge_LongUsername(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	longUsername := strings.Repeat("a", 255) // Assuming 255 is DB limit or reasonably large
	user, password := createGuardianTestUser("password123")
	user.Username = longUsername
	loginReq := model.LoginRequest{Username: longUsername, Password: password}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, longUsername).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, longUsername).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, longUsername).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{"role:user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.wsManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return()

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	})).Return(nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Equal(t, longUsername, loginResp.User.Username)
	deps.userRepo.AssertExpectations(t)
}

// TestAuthUseCase_Vulnerability_SQLInjectionInUsername tests that SQL injection payloads are treated as normal strings by the UseCase logic (Repositories should handle the safety).
func TestAuthUseCase_Vulnerability_SQLInjectionInUsername(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	sqlInjectionUsername := "admin' OR 1=1 --"
	loginReq := model.LoginRequest{Username: sqlInjectionUsername, Password: "password123"}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, sqlInjectionUsername).Return(false, time.Duration(0), nil)

	// In the UseCase, we expect FindByUsername to be called with the raw string.
	// The Repository is responsible for sanitization/parameterization.
	// We simulate a "User Not Found" or "Invalid Credentials" because a real DB wouldn't find this user.
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)

	// Mocking that the repo returns error, effectively saying "no such user found even with injection attempt"
	deps.userRepo.On("FindByUsername", mock.Anything, sqlInjectionUsername).Return(nil, errors.New("record not found"))

	_, _, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials))
	deps.userRepo.AssertExpectations(t)
}

// TestAuthUseCase_Failure_GenerateAndStoreTokenPairError tests error handling when token generation fails (e.g., UUID failure).
func TestAuthUseCase_Failure_GenerateAndStoreTokenPairError(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	user, password := createGuardianTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{"role:user"}, nil)

	// FORCE ERROR HERE
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(errors.New("redis store failed"))

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "failed to store session")
	deps.tokenRepo.AssertExpectations(t)
}

// TestAuthUseCase_Login_AccountLockingLogic tests the locking mechanism more granularly.
func TestAuthUseCase_Login_AccountLockingLogic(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	user, _ := createGuardianTestUser("password123")
	wrongPassword := "wrongpass"
	loginReq := model.LoginRequest{Username: user.Username, Password: wrongPassword}

	// Case 1: Attempts < Max
	t.Run("Increment attempts but do not lock", func(t *testing.T) {
		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil).Once()

		// UseCase calls tm.WithinTransaction
		deps.tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials).Once()

		deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil).Once()

		// IncrementLoginAttempts returns the new attempt count. Return 1, nil.
		deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil).Once()

		_, _, err := authService.Login(context.Background(), loginReq)
		assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	})

	// Case 2: Attempts >= Max
	t.Run("Increment attempts and lock account", func(t *testing.T) {
		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil).Once()

		deps.tm.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrAccountLocked).Once()

		deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil).Once()

		// IncrementLoginAttempts returns 5 (Max), nil.
		deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(5, nil).Once()

		deps.tokenRepo.On("LockAccount", mock.Anything, user.Username, mock.Anything).Return(nil).Once()

		deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "ACCOUNT_LOCKED"
		})).Return(nil).Once()

		_, _, err := authService.Login(context.Background(), loginReq)
		assert.ErrorIs(t, err, usecase.ErrAccountLocked)
	})
}

// TestAuthUseCase_ResetPassword_Edge_LongPassword tests that bcrypt failure is handled.
func TestAuthUseCase_ResetPassword_Edge_LongPassword(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	token := "valid-token"
	resetToken := &authEntity.PasswordResetToken{
		Email:     "user@example.com",
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	// Password longer than 72 bytes causes bcrypt to fail
	longPassword := strings.Repeat("a", 73)

	deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, resetToken.Email).Return(&userEntity.User{Email: resetToken.Email}, nil)

	err := authService.ResetPassword(context.Background(), token, longPassword)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to hash password")
	assert.Contains(t, err.Error(), "bcrypt: password length exceeds 72 bytes")
}

// TestAuthUseCase_ForgotPassword_Edge_EmailDistributorFailure tests graceful degradation.
func TestAuthUseCase_ForgotPassword_Edge_EmailDistributorFailure(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	email := "user@example.com"
	user := &userEntity.User{ID: "user-id", Email: email}

	deps.userRepo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)

	// Mock distributor failure
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("queue error"))

	// Audit log should still happen
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "FORGOT_PASSWORD_REQUEST"
	})).Return(nil)

	err := authService.ForgotPassword(context.Background(), email)

	assert.NoError(t, err) // Should not return error
	deps.taskDistributor.AssertExpectations(t)
}
