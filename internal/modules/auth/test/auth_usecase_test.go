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
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
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

// initializeValidator creates and configures a validator with custom validations
func initializeValidator() *validator.Validate {
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return v
}

type testDependencies struct {
	jwtManager      *jwt.JWTManager
	tokenRepo       *mock_auth.MockTokenRepository
	userRepo        *mock_user.MockUserRepository
	tm              *mocking.MockWithTransactionManager
	wsManager       *mocking.MockManager
	enforcer        *mock_permission.IEnforcer
	validate        *validator.Validate
	log             *logrus.Logger
	auditUC         *auditMocks.MockAuditUseCase
	taskDistributor *mocking.MockTaskDistributor
}

func setupTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	jwtManager := jwt.NewJWTManager(TestAccessSecret, TestRefreshSecret, 15*time.Minute, 24*time.Hour)

	deps := &testDependencies{
		jwtManager:      jwtManager,
		tokenRepo:       new(mock_auth.MockTokenRepository),
		userRepo:        new(mock_user.MockUserRepository),
		tm:              new(mocking.MockWithTransactionManager),
		wsManager:       new(mocking.MockManager),
		enforcer:        new(mock_permission.IEnforcer),
		validate:        initializeValidator(),
		log:             logrus.New(),
		auditUC:         new(auditMocks.MockAuditUseCase),
		taskDistributor: new(mocking.MockTaskDistributor),
	}

	deps.log.SetOutput(io.Discard)

	authService := usecase.NewAuthUsecase(
		deps.jwtManager,
		deps.tokenRepo,
		deps.userRepo,
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
	deps.auditUC.AssertExpectations(t)
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
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials)) // Use errors.Is
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
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
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials)) // Use errors.Is
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
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
	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.enforcer.AssertExpectations(t)
}

func TestLogin_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.wsManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return()

	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	deps.auditUC.AssertExpectations(t)
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

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == "session-1"
	})).Return(nil)

	tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), oldRefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.enforcer.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
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
	authService, _ := setupTest(t)

	_, _, err := authService.RefreshToken(context.Background(), "this.is.an.invalid.token")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
}

func TestRefreshToken_Failure_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	refreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: refreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(nil, gorm.ErrRecordNotFound)

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
}

func TestValidateAccessToken_Failure_Expired(t *testing.T) {
	authService, _ := setupTest(t)

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
}

func TestValidateAccessToken_Failure_Mismatch(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestAccessSecret, 15*time.Minute)
	assert.NoError(t, err)

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
	deps.auditUC.AssertExpectations(t)
}

func TestRevokeToken_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"

	deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.NoError(t, err) // Should proceed
	deps.tokenRepo.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
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

func TestRevokeAllSessions_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	userID := "user-1"

	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, userID).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.RevokeAllSessions(context.Background(), userID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
}

func TestGenerateAccessToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)

	token, err := authService.GenerateAccessToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateAccessToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	_, err := authService.GenerateAccessToken(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	deps.enforcer.AssertExpectations(t)
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{TestRole}, nil)

	token, err := authService.GenerateRefreshToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID).Return(nil, errors.New("casbin error"))

	_, err := authService.GenerateRefreshToken(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	deps.enforcer.AssertExpectations(t)
}

// --- FORGOT & RESET PASSWORD TESTS (Updated with Background Worker) ---

func TestForgotPassword_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(payload *tasks.SendEmailPayload) bool {
		return payload.To == user.Email && payload.Subject == "Password Reset Request"
	}), mock.Anything).Return(nil)

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "FORGOT_PASSWORD_REQUEST"
	})).Return(nil)

	err := authService.ForgotPassword(context.Background(), user.Email)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

func TestForgotPassword_UserNotFound_Security_EnumPrevention(t *testing.T) {
	authService, deps := setupTest(t)
	email := "notfound@example.com"

	deps.userRepo.On("FindByEmail", mock.Anything, email).Return(nil, errors.New("user not found"))

	err := authService.ForgotPassword(context.Background(), email)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything)
}

func TestForgotPassword_Failure_RepositoryError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	// Simulate DB error when saving token
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(errors.New("db save error"))

	err := authService.ForgotPassword(context.Background(), user.Email)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db save error")
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	// Audit log should NOT be called if save fails
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
}

func TestForgotPassword_DistributeTaskError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("task queue error"))

	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := authService.ForgotPassword(context.Background(), user.Email)

	// We expect NO error because email sending failure is logged but doesn't fail the request
	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

func TestForgotPassword_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.ForgotPassword(context.Background(), user.Email)

	assert.NoError(t, err)
}

func TestResetPassword_Success(t *testing.T) {
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
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "PASSWORD_RESET_SUCCESS"
	})).Return(nil)

	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

func TestResetPassword_Failure_TransactionError(t *testing.T) {
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

	// Simulate Transaction Failure
	dbErr := errors.New("update failed")
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(dbErr)

	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(dbErr)

	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
}

func TestResetPassword_Failure_InvalidToken(t *testing.T) {
	authService, deps := setupTest(t)
	token := "invalid-token"

	deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(nil, errors.New("token not found"))

	err := authService.ResetPassword(context.Background(), token, "new-password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidResetToken))
	deps.tokenRepo.AssertExpectations(t)
}

func TestResetPassword_Failure_ExpiredToken(t *testing.T) {
	authService, deps := setupTest(t)
	token := "expired-token"
	resetToken := &authEntity.PasswordResetToken{
		Email:     "test@example.com",
		Token:     token,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.tokenRepo.On("DeleteByEmail", mock.Anything, resetToken.Email).Return(nil)

	err := authService.ResetPassword(context.Background(), token, "new-password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidResetToken))
	deps.tokenRepo.AssertExpectations(t)
}

func TestResetPassword_Failure_UserDeleted(t *testing.T) {
	authService, deps := setupTest(t)
	token := "valid-token"
	resetToken := &authEntity.PasswordResetToken{
		Email:     "deleted@example.com",
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, resetToken.Email).Return(nil, errors.New("user not found"))

	err := authService.ResetPassword(context.Background(), token, "new-password")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidResetToken))
	deps.userRepo.AssertExpectations(t)
}

func TestResetPassword_AuditError(t *testing.T) {
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
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.NoError(t, err)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: token}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)

	claims, err := authService.ValidateRefreshToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "session-1", claims.SessionID)
	deps.tokenRepo.AssertExpectations(t)
}

func TestValidateRefreshToken_Failure_Expired(t *testing.T) {
	authService, _ := setupTest(t)

	token, err := jwt.GenerateTestToken(TestUserID, "session-1", TestRole, TestUsername, TestRefreshSecret, -1*time.Hour)
	assert.NoError(t, err)

	claims, err := authService.ValidateRefreshToken(token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
	assert.Nil(t, claims)
}

func TestValidateRefreshToken_Failure_Revoked(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	// Simulate token not found in Redis (revoked)
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(nil, nil)

	claims, err := authService.ValidateRefreshToken(token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrTokenRevoked))
	assert.Nil(t, claims)
	deps.tokenRepo.AssertExpectations(t)
}

func TestLogin_Success_NoRoles(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	// Return empty roles
	deps.enforcer.On("GetRolesForUser", user.ID).Return([]string{}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.wsManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return()

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	})).Return(nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Empty(t, loginResp.User.Role)
	deps.enforcer.AssertExpectations(t)
}
