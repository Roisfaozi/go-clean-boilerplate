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
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
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
	jwtManager      *jwt.JWTManager
	tokenRepo       *mock_auth.MockTokenRepository
	userRepo        *mock_user.MockUserRepository
	orgRepo         *mock_org.MockOrganizationRepository
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
		orgRepo:         new(mock_org.MockOrganizationRepository),
		tm:              new(mocking.MockWithTransactionManager),
		wsManager:       new(mocking.MockManager),
		enforcer:        new(mock_permission.IEnforcer),
		validate:        validator.New(),
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
		deps.orgRepo,
		deps.tm,
		deps.log,
		deps.wsManager,
		(*sse.Manager)(nil),
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
		Status:   entity.UserStatusActive,
	}, password
}

func TestLogin_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "127.0.0.1", UserAgent: "TestAgent"}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

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

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, "nonexistent").Return(false, time.Duration(0), nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)
	deps.userRepo.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials))
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
}

func TestLogin_Failure_InvalidPassword(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: "wrong-password"}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidCredentials))
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
}

func TestLogin_Failure_StoreTokenError(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}
	storeErr := errors.New("redis is down")

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{TestRole}, nil)
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

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return(nil, errors.New("casbin error"))

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

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(errors.New("audit error"))

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	deps.auditUC.AssertExpectations(t)
}

func TestLogin_Security_BruteForceProtection(t *testing.T) {

	maxAttempts := 1
	lockoutDuration := 15 * time.Minute

	_, deps := setupTest(t)
	authService := usecase.NewAuthUsecase(
		maxAttempts,
		lockoutDuration,
		deps.jwtManager,
		deps.tokenRepo,
		deps.userRepo,
		deps.orgRepo,
		deps.tm,
		deps.log,
		deps.wsManager,
		nil,
		deps.enforcer,
		deps.auditUC,
		deps.taskDistributor,
	)

	user, _ := createTestUser("password123")
	wrongPassword := "wrongpass"

	t.Run("Should lock account immediately if max attempts reached", func(t *testing.T) {

		deps.tm.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)

		deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

		deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil)

		deps.tokenRepo.On("LockAccount", mock.Anything, user.Username, lockoutDuration).Return(nil)

		deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

		req := model.LoginRequest{
			Username: user.Username,
			Password: wrongPassword,
		}

		_, _, err := authService.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many failed attempts")
	})
}

func TestRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{TestRole}, nil)
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
	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return(nil, errors.New("casbin error"))

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

	assert.NoError(t, err)
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

	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{TestRole}, nil)

	token, err := authService.GenerateAccessToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateAccessToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return(nil, errors.New("casbin error"))

	_, err := authService.GenerateAccessToken(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	deps.enforcer.AssertExpectations(t)
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{TestRole}, nil)

	token, err := authService.GenerateRefreshToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return(nil, errors.New("casbin error"))

	_, err := authService.GenerateRefreshToken(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	deps.enforcer.AssertExpectations(t)
}

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

	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(errors.New("db save error"))

	err := authService.ForgotPassword(context.Background(), user.Email)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)

	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
}

func TestLogin_Failure_UserSuspended(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	user.Status = entity.UserStatusSuspended

	loginReq := model.LoginRequest{Username: user.Username, Password: password}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrAccountSuspended)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAccountSuspended))
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.tokenRepo.AssertExpectations(t)
}

func TestRefreshToken_Failure_UserSuspended(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	user.Status = entity.UserStatusBanned

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: token}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)

	_, _, err = authService.RefreshToken(context.Background(), token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAccountSuspended))
}

func TestForgotPassword_DistributeTaskError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("task queue error"))

	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := authService.ForgotPassword(context.Background(), user.Email)

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

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "REVOKE_ALL_SESSIONS"
	})).Return(nil)

	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)

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

	// Revoke sessions
	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)

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

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

	deps.enforcer.On("GetRolesForUser", user.ID, "global").Return([]string{}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	})).Return(nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Empty(t, loginResp.User.Role)
	deps.enforcer.AssertExpectations(t)
}

func TestRequestVerification_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil

	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(payload *tasks.SendEmailPayload) bool {
		return payload.To == user.Email && payload.Subject == "Verify Your Email Address"
	}), mock.Anything).Return(nil)

	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "VERIFICATION_EMAIL_REQUESTED"
	})).Return(nil)

	err := authService.RequestVerification(context.Background(), user.ID)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

func TestRequestVerification_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)

	deps.userRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, errors.New("user not found"))

	err := authService.RequestVerification(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	deps.userRepo.AssertExpectations(t)
}

func TestRequestVerification_AlreadyVerified(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	verifiedAt := time.Now().UnixMilli()
	user.EmailVerifiedAt = &verifiedAt

	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)

	err := authService.RequestVerification(context.Background(), user.ID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAlreadyVerified))
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertNotCalled(t, "SaveVerificationToken", mock.Anything, mock.Anything)
}

func TestRequestVerification_SaveTokenError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil

	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(errors.New("db error"))

	err := authService.RequestVerification(context.Background(), user.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	deps.tokenRepo.AssertExpectations(t)
}

func TestRequestVerification_DistributeTaskError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil

	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("task queue error"))
	deps.auditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := authService.RequestVerification(context.Background(), user.ID)

	assert.NoError(t, err)
	deps.taskDistributor.AssertExpectations(t)
}

func TestVerifyEmail_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil
	token := "valid-verification-token"
	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: now + (24 * 60 * 60 * 1000),
		CreatedAt: now,
	}

	deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verificationToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "EMAIL_VERIFIED"
	})).Return(nil)

	err := authService.VerifyEmail(context.Background(), token)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.auditUC.AssertExpectations(t)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	authService, deps := setupTest(t)
	token := "invalid-token"

	deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(nil, errors.New("not found"))

	err := authService.VerifyEmail(context.Background(), token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidVerificationToken))
	deps.tokenRepo.AssertExpectations(t)
}

func TestVerifyEmail_ExpiredToken(t *testing.T) {
	authService, deps := setupTest(t)
	token := "expired-token"
	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     "test@example.com",
		Token:     token,
		ExpiresAt: now - (1 * 60 * 60 * 1000),
		CreatedAt: now - (25 * 60 * 60 * 1000),
	}

	deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verificationToken, nil)
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, verificationToken.Email).Return(nil)

	err := authService.VerifyEmail(context.Background(), token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidVerificationToken))
	deps.tokenRepo.AssertExpectations(t)
}

func TestVerifyEmail_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	token := "valid-token"
	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     "deleted@example.com",
		Token:     token,
		ExpiresAt: now + (24 * 60 * 60 * 1000),
		CreatedAt: now,
	}

	deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verificationToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, verificationToken.Email).Return(nil, errors.New("user not found"))

	err := authService.VerifyEmail(context.Background(), token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidVerificationToken))
	deps.userRepo.AssertExpectations(t)
}

func TestVerifyEmail_AlreadyVerified(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	verifiedAt := time.Now().UnixMilli()
	user.EmailVerifiedAt = &verifiedAt
	token := "valid-token"
	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: now + (24 * 60 * 60 * 1000),
		CreatedAt: now,
	}

	deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verificationToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil)

	err := authService.VerifyEmail(context.Background(), token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrAlreadyVerified))
	deps.tokenRepo.AssertExpectations(t)
}

func TestVerifyEmail_TransactionError(t *testing.T) {
	authService, deps := setupTest(t)
	_ = authService // Prevent unused warning if test fails early
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil
	token := "valid-token"
	now := time.Now().UnixMilli()
	verificationToken := &authEntity.EmailVerificationToken{
		Email:     user.Email,
		Token:     token,
		ExpiresAt: now + (24 * 60 * 60 * 1000),
		CreatedAt: now,
	}

	deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verificationToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)

	dbErr := errors.New("update failed")
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(dbErr)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(dbErr)

	err := authService.VerifyEmail(context.Background(), token)

	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	deps.auditUC.AssertNotCalled(t, "LogActivity", mock.Anything, mock.Anything)
}

func TestRegister_Success(t *testing.T) {
	authService, deps := setupTest(t)
	password := "password123"
	req := model.RegisterRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  password,
		Name:      "New User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPassword := string(hashedBytes)

	// 1. Check existing (Register check)
	deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(nil, gorm.ErrRecordNotFound).Once()
	deps.userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, gorm.ErrRecordNotFound)

	// 2. Transaction
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)

	// In Transaction:
	// Create User
	deps.userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Username == req.Username && u.Email == req.Email
	})).Return(nil)

	// Add Role
	deps.enforcer.On("AddGroupingPolicy", mock.Anything, "role:user", "global").Return(true, nil)

	// Create Org (Auto-Provisioning)
	deps.orgRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *orgEntity.Organization) bool {
		return o.Name == "New User's Workspace"
	}), "owner").Return(nil)

	// Audit (Register Action)
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "REGISTER" && req.Entity == "User"
	})).Return(nil)

	// 4. Login (Implicitly called by Register)
	// Login logic mocks:
	deps.tokenRepo.On("IsAccountLocked", mock.Anything, req.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, req.Username).Return(nil)

	// FindByUsername for Login (Second call) - MUST RETURN USER with matching password
	createdUser := &entity.User{
		ID:       "new-user-id",
		Username: req.Username,
		Password: hashedPassword,
		Status:   entity.UserStatusActive,
	}
	deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(createdUser, nil).Once()

	// Enforcer GetRolesForUser (Login)
	deps.enforcer.On("GetRolesForUser", createdUser.ID, "global").Return([]string{"role:user"}, nil)

	// StoreToken
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

	// FindUserOrganizations (Login)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, createdUser.ID).Return([]*orgEntity.Organization{}, nil)

	// Audit Login
	deps.auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "LOGIN"
	})).Return(nil)

	// Execute
	loginResp, refreshToken, err := authService.Register(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, req.Username, loginResp.User.Username)
	assert.Equal(t, "new-user-id", loginResp.User.ID)

	deps.userRepo.AssertExpectations(t)
	deps.orgRepo.AssertExpectations(t)
	deps.enforcer.AssertExpectations(t)
}

func TestRegister_Fail_UsernameExists(t *testing.T) {
	authService, deps := setupTest(t)
	req := model.RegisterRequest{
		Username: "existing",
		Email:    "new@example.com",
		Password: "password123",
	}

	deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(&entity.User{ID: "existing-id"}, nil)

	loginResp, refreshToken, err := authService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username already exists")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
}

func TestRegister_Fail_EmailExists(t *testing.T) {
	authService, deps := setupTest(t)
	req := model.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(nil, gorm.ErrRecordNotFound)
	deps.userRepo.On("FindByEmail", mock.Anything, req.Email).Return(&entity.User{ID: "existing-id"}, nil)

	loginResp, refreshToken, err := authService.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
}
