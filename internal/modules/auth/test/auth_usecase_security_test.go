package test

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	mock_permission "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// SECURITY TEST SUITE - Auth UseCase
// Tests for: Race conditions, Token replay, Session hijacking, Concurrent access
// ============================================================================

type securityTestDeps struct {
	jwtManager      *jwt.JWTManager
	tokenRepo       *mock_auth.MockTokenRepository
	userRepo        *mock_user.MockUserRepository
	orgRepo         *mock_org.MockOrganizationRepository
	tm              *mocking.MockWithTransactionManager
	wsManager       *mocking.MockManager
	enforcer        *mock_permission.IEnforcer
	log             *logrus.Logger
	auditUC         *auditMocks.MockAuditUseCase
	taskDistributor *mocking.MockTaskDistributor
}

func setupSecurityTest(t *testing.T) (usecase.AuthUseCase, *securityTestDeps) {
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)

	deps := &securityTestDeps{
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
		5,              // maxLoginAttempts
		30*time.Minute, // lockoutDuration
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

func createSecurityTestUser(password string) (*userEntity.User, string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &userEntity.User{
		ID:       "user-security-test",
		Username: "securityuser",
		Name:     "Security Test User",
		Password: string(hashedPassword),
		Email:    "security@example.com",
		Status:   userEntity.UserStatusActive,
	}, password
}

// ============================================================================
// 🔐 CONCURRENT LOGIN ATTEMPT TESTS
// ============================================================================

// TestLogin_Concurrent_MultipleFailedAttempts tests race condition when incrementing login attempts.
func TestLogin_Concurrent_MultipleFailedAttempts(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	wrongPassword := "wrongpassword"
	numConcurrent := 10

	var attemptCounter int32

	// Setup mocks for concurrent access
	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)

	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

	// Counter to track concurrent increments
	deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&attemptCounter, 1)
		}).Return(3, nil) // Return a value less than max to avoid account locking

	var wg sync.WaitGroup
	errChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loginReq := model.LoginRequest{
				Username: user.Username,
				Password: wrongPassword,
			}
			_, _, err := authService.Login(context.Background(), loginReq)
			errChan <- err
		}()
	}

	wg.Wait()
	close(errChan)

	// Verify all attempts received ErrInvalidCredentials
	for err := range errChan {
		assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	}

	// Verify increment was called for each attempt
	assert.Equal(t, int32(numConcurrent), atomic.LoadInt32(&attemptCounter), "IncrementLoginAttempts should be called for each failed attempt")
}

// TestLogin_Concurrent_AccountLockAtThreshold tests that account locks exactly at max attempts.
func TestLogin_Concurrent_AccountLockAtThreshold(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")

	var lockCalled int32

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil).Once()

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrAccountLocked).Once()

	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil).Once()

	// Return 5 (which equals max) to trigger lock
	deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(5, nil).Once()

	deps.tokenRepo.On("LockAccount", mock.Anything, user.Username, mock.Anything).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&lockCalled, 1)
		}).Return(nil).Once()

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "ACCOUNT_LOCKED"
	})).Return(nil).Once()

	loginReq := model.LoginRequest{
		Username: user.Username,
		Password: "wrongpassword",
	}
	_, _, err := authService.Login(context.Background(), loginReq)

	assert.ErrorIs(t, err, usecase.ErrAccountLocked)
	assert.Equal(t, int32(1), atomic.LoadInt32(&lockCalled), "LockAccount should be called exactly once at threshold")
}

// ============================================================================
// 🔐 SESSION CLEANUP EDGE CASES
// ============================================================================

// TestRefreshToken_SessionCleanupFailure tests that refresh proceeds even if old session revocation fails.
func TestRefreshToken_SessionCleanupFailure(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	sessionID := "session-to-refresh"

	// Generate a valid refresh token
	refreshToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "test-refresh-secret", 24*time.Hour)
	assert.NoError(t, err)

	// Mock session valid (for ValidateRefreshToken -> validateSession)
	savedSession := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		AccessToken:  "old-access-token",
	}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(savedSession, nil)

	// Mock user lookup
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)

	// Mock enforcer
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{"role:user"}, nil)

	// Mock RevokeToken internal calls:
	// 1. AuditUC.LogActivity (LOGOUT action)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "LOGOUT" && req.Entity == "Auth"
	})).Return(nil)

	// 2. FORCE ERROR: Old session deletion fails
	deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, sessionID).Return(errors.New("redis connection lost"))

	// But new session should still be created (generateAndStoreTokenPair)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

	tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), refreshToken)

	// Should succeed despite cleanup failure (graceful degradation)
	assert.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, newRefreshToken)
}

// TestRefreshToken_OrphanedSession tests behavior with expired but not-yet-deleted session.
func TestRefreshToken_ExpiredButOrphanedSession(t *testing.T) {
	authService, _ := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	sessionID := "orphaned-session"

	// Generate an EXPIRED refresh token
	expiredToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "test-refresh-secret", -1*time.Hour)
	assert.NoError(t, err)

	// No need to mock anything else - JWT validation should fail first
	_, _, err = authService.RefreshToken(context.Background(), expiredToken)

	assert.ErrorIs(t, err, usecase.ErrInvalidToken)
}

// ============================================================================
// 🔐 TOKEN REPLAY ATTACK TESTS
// ============================================================================

// TestValidateAccessToken_ReplayAttack_SameTokenAfterRefresh tests using old access token after refresh.
func TestValidateAccessToken_ReplayAttack_SameTokenAfterRefresh(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	sessionID := "session-1"

	// Generate old access token
	oldAccessToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "test-access-secret", 15*time.Minute)
	assert.NoError(t, err)

	// After refresh, stored token is NEW, but attacker uses OLD token
	newSession := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  "new-access-token-after-refresh",
		RefreshToken: "new-refresh-token",
	}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(newSession, nil)

	// Attempt to use old token should fail (token mismatch)
	claims, err := authService.ValidateAccessToken(oldAccessToken)

	assert.ErrorIs(t, err, usecase.ErrTokenRevoked)
	assert.Nil(t, claims)
}

// TestValidateRefreshToken_ReplayAttack_SameTokenAfterRefresh tests using old refresh token after refresh.
func TestValidateRefreshToken_ReplayAttack_SameTokenAfterRefresh(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	sessionID := "session-1"

	// Generate old refresh token
	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "test-refresh-secret", 24*time.Hour)
	assert.NoError(t, err)

	// After refresh, stored token is NEW
	newSession := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token-after-refresh",
	}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(newSession, nil)

	// Attempt to use old refresh token should fail
	claims, err := authService.ValidateRefreshToken(oldRefreshToken)

	assert.ErrorIs(t, err, usecase.ErrTokenRevoked)
	assert.Nil(t, claims)
}

// ============================================================================
// 🔐 VERIFICATION TOKEN REPLAY TESTS
// ============================================================================

// TestVerifyEmail_TokenReplay_SameTokenTwice tests using same verification token twice.
func TestVerifyEmail_TokenReplay_SameTokenTwice(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	user.EmailVerifiedAt = nil // Not verified initially
	verificationToken := "verification-token-123"

	// First verification - token exists and is valid
	tokenEntity := &authEntity.EmailVerificationToken{
		Email:     user.Email,
		Token:     verificationToken,
		ExpiresAt: time.Now().Add(15 * time.Minute).UnixMilli(), // Use UnixMilli as per implementation
	}

	// Mock: Find verification token (first time - token exists)
	deps.tokenRepo.On("FindVerificationToken", mock.Anything, verificationToken).Return(tokenEntity, nil).Once()

	// Mock: Find user by email
	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil).Once()

	// Mock: Transaction for update
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil).Once()

	// Mock: Update user (sets EmailVerifiedAt)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil).Once()

	// Mock: Delete verification token by email
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil).Once()

	// Mock: Audit log
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "EMAIL_VERIFIED" && req.Entity == "User"
	})).Return(nil).Once()

	// First call should succeed
	err := authService.VerifyEmail(context.Background(), verificationToken)
	assert.NoError(t, err)

	// Second call with same token - token should be deleted/not found
	deps.tokenRepo.On("FindVerificationToken", mock.Anything, verificationToken).Return(nil, errors.New("not found")).Once()

	err = authService.VerifyEmail(context.Background(), verificationToken)
	assert.ErrorIs(t, err, usecase.ErrInvalidVerificationToken)
}

// ============================================================================
// 🔐 CONCURRENT SESSION VALIDATION TESTS
// ============================================================================

// TestValidateAccessToken_Concurrent_MultipleGoroutines tests concurrent token validation.
func TestValidateAccessToken_Concurrent_MultipleGoroutines(t *testing.T) {
	authService, deps := setupSecurityTest(t)
	user, _ := createSecurityTestUser("password123")
	sessionID := "concurrent-session"
	numConcurrent := 20

	// Generate valid access token
	accessToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "test-access-secret", 15*time.Minute)
	assert.NoError(t, err)

	// Mock valid session
	validSession := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: "refresh-token",
	}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(validSession, nil)

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claims, err := authService.ValidateAccessToken(accessToken)
			if err == nil && claims != nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}()
	}

	wg.Wait()

	// All validations should succeed (no race condition causing failures)
	assert.Equal(t, int32(numConcurrent), successCount, "All concurrent validations should succeed")
	assert.Equal(t, int32(0), failCount, "No failures expected in concurrent validation")
}

// ============================================================================
// 🔐 NIL DEPENDENCY HANDLING
// ============================================================================

// TestLogin_NilEnforcer tests login when Enforcer is nil (RBAC disabled).
func TestLogin_NilEnforcer(t *testing.T) {
	// Create service with nil enforcer
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)

	tokenRepo := new(mock_auth.MockTokenRepository)
	userRepo := new(mock_user.MockUserRepository)
	orgRepo := new(mock_org.MockOrganizationRepository)
	tm := new(mocking.MockWithTransactionManager)
	wsManager := new(mocking.MockManager)
	auditUC := new(auditMocks.MockAuditUseCase)
	taskDistributor := new(mocking.MockTaskDistributor)

	log := logrus.New()
	log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		log,
		wsManager,
		nil, // sseManager
		nil, // NIL ENFORCER
		auditUC,
		taskDistributor,
	)

	user, password := createSecurityTestUser("password123")

	tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)

	userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)
	auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	loginReq := model.LoginRequest{Username: user.Username, Password: password}
	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, refreshToken)
	// Role should be empty when Enforcer is nil
	assert.Empty(t, loginResp.User.Role)
}

// TestLogin_NilAuditUC tests login when AuditUC is nil.
func TestLogin_NilAuditUC(t *testing.T) {
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)

	tokenRepo := new(mock_auth.MockTokenRepository)
	userRepo := new(mock_user.MockUserRepository)
	orgRepo := new(mock_org.MockOrganizationRepository)
	tm := new(mocking.MockWithTransactionManager)
	wsManager := new(mocking.MockManager)
	enforcer := new(mock_permission.IEnforcer)
	taskDistributor := new(mocking.MockTaskDistributor)

	log := logrus.New()
	log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		log,
		wsManager,
		nil,
		enforcer,
		nil, // NIL AUDIT UC
		taskDistributor,
	)

	user, password := createSecurityTestUser("password123")

	tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)

	userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{"role:user"}, nil)
	tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	loginReq := model.LoginRequest{Username: user.Username, Password: password}
	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	// Should succeed even without audit logging
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, refreshToken)
}
