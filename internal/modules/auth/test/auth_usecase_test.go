package test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	mock_permission "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	TestAccessSecret  = "test-access-secret"
	TestRefreshSecret = "test-refresh-secret"
	TestUserID        = "user-test-id"
	TestUsername      = "testuser"
	TestRole          = "role:user"
)

type testDependencies struct {
	jwtManager *jwt.JWTManager
	tokenRepo  *mock_auth.MockTokenRepository
	userRepo   *mock_user.MockUserRepository
	tm         *mocking.MockWithTransactionManager
	wsManager  *mocking.MockManager
	enforcer   *mock_permission.IEnforcer
	validate   *validator.Validate
	log        *logrus.Logger
	auditUC    *auditMocks.MockAuditUseCase // Added MockAuditUseCase
}

func setupTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	jwtManager := jwt.NewJWTManager(TestAccessSecret, TestRefreshSecret, 15*time.Minute, 24*time.Hour)

	deps := &testDependencies{
		jwtManager: jwtManager,
		tokenRepo:  new(mock_auth.MockTokenRepository),
		userRepo:   new(mock_user.MockUserRepository),
		tm:         new(mocking.MockWithTransactionManager),
		wsManager:  new(mocking.MockManager),
		enforcer:   new(mock_permission.IEnforcer),
		validate:   validator.New(),
		log:        logrus.New(),
		auditUC:    new(auditMocks.MockAuditUseCase), // Initialize mock audit usecase
	}

	deps.log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		deps.jwtManager,
		deps.tokenRepo,
		deps.userRepo,
		deps.tm,
		deps.log,
		deps.wsManager,
		deps.enforcer,
		deps.auditUC, // Pass mock audit usecase
	)

	return authService, deps
}

func createTestUser(password string) (*entity.User, string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &entity.User{
		ID:       TestUserID,
		Username: TestUsername,
		Name:     "Test User",
		Password: string(hashedPassword),
		Email:    "test@example.com",
	}, password
}

func TestLogin_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "127.0.0.1", UserAgent: "TestAgent"}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.wsManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return()
	
	// Expect LogActivity with a Matcher
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN" && req.Entity == "Auth" && req.IPAddress == loginReq.IPAddress
	})).Return(nil)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, user.ID, loginResp.User.ID)
	assert.Equal(t, user.Username, loginResp.User.Username)
	assert.Equal(t, TestRole, loginResp.User.Role)

	deps.tm.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.enforcer.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.wsManager.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t) // Assert audit log call
}

func TestLogin_Failure_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	loginReq := model.LoginRequest{Username: "nonexistent", Password: "password123"}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)
	deps.userRepo.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything) // Should not log on failure
}

func TestLogin_Failure_InvalidPassword(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: "wrong-password"}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything) // Should not log on failure
}

func TestLogin_Failure_StoreTokenError(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}
	storeErr := errors.New("redis is down")

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(storeErr)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store session")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything) // Should not log on StoreToken failure
}

func TestLogin_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	// Enforcer returns error
	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.enforcer.AssertExpectations(t)
}


func TestRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)
	deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, "session-1").Return(nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	
	// Expect LogActivity because RevokeToken is called
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == "session-1" // sessionID needs to be verified
	})).Return(nil) // audit log is called by RevokeToken

	tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), oldRefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.enforcer.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t) // Audit log for RevokeToken within RefreshToken flow
}

func TestRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	_, _, err = authService.RefreshToken(context.Background(), oldRefreshToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
}


func TestRefreshToken_Failure_InvalidToken(t *testing.T) {
	authService, deps := setupTest(t) // Always declare both

	_, _, err := authService.RefreshToken(context.Background(), "this.is.an.invalid.token")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
}

func TestRefreshToken_Failure_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	refreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: refreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(nil, gorm.ErrRecordNotFound)
	
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything) // Audit log should not be called if user not found before RevokeToken is called

	_, _, err = authService.RefreshToken(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	deps.userRepo.AssertExpectations(t)
}

func TestValidateAccessToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestAccessSecret, 15*time.Minute)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, AccessToken: token}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)

	claims, err := authService.ValidateAccessToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "session-1", claims.SessionID)
	assert.Equal(t, TestRole, claims.Role)
	assert.Equal(t, user.Username, claims.Username)
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything) // No audit log for validation
}

func TestValidateAccessToken_Failure_Expired(t *testing.T) {
	authService, _ := setupTest(t) // Fix unused
	// Removed the shadowing `_ = authService` since the initial declaration is now `authService, _`.

	expiredToken, err := jwt.GenerateTestToken("user-id", "session-1", TestRole, TestUsername, TestAccessSecret, -1*time.Hour)
	assert.NoError(t, err)

	claims, err := authService.ValidateAccessToken(expiredToken)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
	assert.Nil(t, claims)
}

func TestValidateAccessToken_Failure_TokenRevoked(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestAccessSecret, 15*time.Minute)
	assert.NoError(t, err)

	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(nil, nil)

	claims, err := authService.ValidateAccessToken(token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrTokenRevoked))
	assert.Nil(t, claims)
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything) // No audit log for validation
}

func TestValidateAccessToken_Failure_Mismatch(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestAccessSecret, 15*time.Minute)
	assert.NoError(t, err)

	// Return a session where access token doesn't match
	session := &model.Auth{ID: "session-1", UserID: user.ID, AccessToken: "different-token"}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)

	claims, err := authService.ValidateAccessToken(token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrTokenRevoked))
	assert.Nil(t, claims)
	deps.tokenRepo.AssertExpectations(t)
}

func TestRevokeToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"

	deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == sessionID
	})).Return(nil)

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t) // Assert audit log call
}

func TestRevokeToken_Failure(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"
	revokeErr := errors.New("failed to delete")

	deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(revokeErr)
	// Audit log should be called even if DeleteToken fails, to log the attempt to revoke.
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == sessionID
	})).Return(nil)

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.Error(t, err)
	assert.Equal(t, revokeErr, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t) // Assert audit log call
}

func TestVerify_Success(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"
	expectedSession := &model.Auth{ID: sessionID, UserID: userID}

	deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(expectedSession, nil)

	session, err := authService.Verify(context.Background(), userID, sessionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSession, session)
	deps.tokenRepo.AssertExpectations(t)
}

func TestGetUserSessions_Success(t *testing.T) {
	authService, deps := setupTest(t)
	userID := "user-1"
	expectedSessions := []*model.Auth{{ID: "session-1", UserID: userID}}

	deps.tokenRepo.On("GetUserSessions", mock.Anything, userID).Return(expectedSessions, nil)

	sessions, err := authService.GetUserSessions(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSessions, sessions)
	deps.tokenRepo.AssertExpectations(t)
}

func TestRevokeAllSessions_Success(t *testing.T) {
	authService, deps := setupTest(t)
	userID := "user-1"

	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, userID).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "REVOKE_ALL_SESSIONS" && req.Entity == "Auth" && req.EntityID == userID
	})).Return(nil)

	err := authService.RevokeAllSessions(context.Background(), userID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

func TestGenerateAccessToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)

	token, err := authService.GenerateAccessToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	deps.enforcer.AssertExpectations(t)
}

func TestGenerateAccessToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	token, err := authService.GenerateAccessToken(user)

	assert.Error(t, err)
	assert.Empty(t, token)
	deps.enforcer.AssertExpectations(t)
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)

	token, err := authService.GenerateRefreshToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	deps.enforcer.AssertExpectations(t)
}

func TestGenerateRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	token, err := authService.GenerateRefreshToken(user)

	assert.Error(t, err)
	assert.Empty(t, token)
	deps.enforcer.AssertExpectations(t)
}
