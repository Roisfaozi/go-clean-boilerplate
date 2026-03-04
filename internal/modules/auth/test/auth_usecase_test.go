package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
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
	publisher       *mock_auth.MockNotificationPublisher
	authz           *mock_auth.MockAuthzManager
	validate        *validator.Validate
	log             *logrus.Logger
	taskDistributor *mocking.MockTaskDistributor
	ticketManager   *mock_auth.MockTicketManager
}

func setupTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	jwtManager := jwt.NewJWTManager(TestAccessSecret, TestRefreshSecret, 15*time.Minute, 24*time.Hour)

	deps := &testDependencies{
		jwtManager:      jwtManager,
		tokenRepo:       new(mock_auth.MockTokenRepository),
		userRepo:        new(mock_user.MockUserRepository),
		orgRepo:         new(mock_org.MockOrganizationRepository),
		tm:              new(mocking.MockWithTransactionManager),
		publisher:       new(mock_auth.MockNotificationPublisher),
		authz:           new(mock_auth.MockAuthzManager),
		validate:        validator.New(),
		log:             logrus.New(),
		taskDistributor: new(mocking.MockTaskDistributor),
		ticketManager:   new(mock_auth.MockTicketManager),
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
		deps.publisher,
		deps.authz,
		deps.taskDistributor,
		deps.ticketManager,
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN" && req.Entity == "Auth" && req.IPAddress == loginReq.IPAddress
	}), mock.Anything).Return(nil)

	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

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
	deps.authz.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.publisher.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.authz.AssertExpectations(t)
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))
	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	deps.taskDistributor.AssertExpectations(t)
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
		deps.publisher,
		deps.authz,
		deps.taskDistributor,
		deps.ticketManager,
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

		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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

	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
	deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, "session-1").Return(nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == "session-1"
	}), mock.Anything).Return(nil)

	tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), oldRefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.authz.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

func TestRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))

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

	refreshToken, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
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

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestAccessSecret, 15*time.Minute)
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

	expiredToken, err := jwt.GenerateTestToken("user-id", "session-1", TestRole, TestUsername, "", TestAccessSecret, -1*time.Hour)
	assert.NoError(t, err)

	claims, err := authService.ValidateAccessToken(expiredToken)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
	assert.Nil(t, claims)
}

func TestValidateAccessToken_Failure_TokenRevoked(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestAccessSecret, 15*time.Minute)
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

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestAccessSecret, 15*time.Minute)
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
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == sessionID
	}), mock.Anything).Return(nil)

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

func TestRevokeToken_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"

	deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
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
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "REVOKE_ALL_SESSIONS" && req.Entity == "Auth" && req.EntityID == userID
	}), mock.Anything).Return(nil)

	err := authService.RevokeAllSessions(context.Background(), userID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

func TestRevokeAllSessions_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	userID := "user-1"

	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, userID).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.RevokeAllSessions(context.Background(), userID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
}

func TestGenerateAccessToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)

	token, err := authService.GenerateAccessToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateAccessToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))

	_, err := authService.GenerateAccessToken(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	deps.authz.AssertExpectations(t)
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)

	token, err := authService.GenerateRefreshToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))

	_, err := authService.GenerateRefreshToken(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	deps.authz.AssertExpectations(t)
}

func TestForgotPassword_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(payload *tasks.SendEmailPayload) bool {
		return payload.To == user.Email && payload.Subject == "Password Reset Request"
	}), mock.Anything).Return(nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "FORGOT_PASSWORD_REQUEST"
	}), mock.Anything).Return(nil)

	err := authService.ForgotPassword(context.Background(), user.Email)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
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

	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything)
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

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
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

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := authService.ForgotPassword(context.Background(), user.Email)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

func TestForgotPassword_AuditError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))

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
	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "PASSWORD_RESET_SUCCESS"
	}), mock.Anything).Return(nil)

	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
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

	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(dbErr)

	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything)
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
	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)
	deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))

	err := authService.ResetPassword(context.Background(), token, "new-strong-password-123")

	assert.NoError(t, err)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
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

	token, err := jwt.GenerateTestToken(TestUserID, "session-1", TestRole, TestUsername, "", TestRefreshSecret, -1*time.Hour)
	assert.NoError(t, err)

	claims, err := authService.ValidateRefreshToken(token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
	assert.Nil(t, claims)
}

func TestValidateRefreshToken_Failure_Revoked(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
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

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)
	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	}), mock.Anything).Return(nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Empty(t, loginResp.User.Role)
	deps.authz.AssertExpectations(t)
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

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "VERIFICATION_EMAIL_REQUESTED"
	}), mock.Anything).Return(nil)

	err := authService.RequestVerification(context.Background(), user.ID)

	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
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
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "EMAIL_VERIFIED"
	}), mock.Anything).Return(nil)

	err := authService.VerifyEmail(context.Background(), token)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
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
	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything)
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

	// Add Role (via AuthzManager)
	deps.authz.On("AssignDefaultRole", mock.Anything, mock.Anything).Return(nil)

	// Create Org (Auto-Provisioning)
	deps.orgRepo.On("Create", mock.Anything, mock.MatchedBy(func(o *orgEntity.Organization) bool {
		return o.Name == "New User's Workspace"
	}), "owner").Return(nil)

	// Audit (Register Action)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "REGISTER" && req.Entity == "User"
	}), mock.Anything).Return(nil)

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

	// Authz GetRolesForUser (Login)
	deps.authz.On("GetRolesForUser", mock.Anything, createdUser.ID, "").Return([]string{"role:user"}, nil)

	// StoreToken
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

	// FindUserOrganizations (Login)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, createdUser.ID).Return([]*orgEntity.Organization{}, nil)

	// Notification
	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	// Audit Login
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "LOGIN"
	}), mock.Anything).Return(nil)

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
	deps.authz.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

// --- Merged from auth_usecase_security_test.go ---
// ============================================================================
// SECURITY TEST SUITE - Auth UseCase
// Tests for: Race conditions, Token replay, Session hijacking, Concurrent access
// ============================================================================

// func createTestUser(password string) (*userEntity.User, string) {
// 	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	return &userEntity.User{
// 		ID:       "user-security-test",
// 		Username: "securityuser",
// 		Name:     "Security Test User",
// 		Password: string(hashedPassword),
// 		Email:    "security@example.com",
// 		Status:   userEntity.UserStatusActive,
// 	}, password
// }

// ============================================================================
// 🔐 CONCURRENT LOGIN ATTEMPT TESTS
// ============================================================================

// TestLogin_Concurrent_MultipleFailedAttempts tests race condition when incrementing login attempts.
func TestLogin_Concurrent_MultipleFailedAttempts(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

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

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "ACCOUNT_LOCKED"
	}), mock.Anything).Return(nil).Once()

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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	sessionID := "session-to-refresh"

	// Generate a valid refresh token
	refreshToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", 24*time.Hour)
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

	// Mock authz
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)

	// Mock RevokeToken internal calls:
	// 1. taskDistributor.DistributeTaskAuditLog (LOGOUT action)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "LOGOUT" && req.Entity == "Auth"
	}), mock.Anything).Return(nil)

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
	authService, _ := setupTest(t)
	user, _ := createTestUser("password123")
	sessionID := "orphaned-session"

	// Generate an EXPIRED refresh token
	expiredToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", -1*time.Hour)
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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	sessionID := "session-1"

	// Generate old access token
	oldAccessToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-access-secret", 15*time.Minute)
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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	sessionID := "session-1"

	// Generate old refresh token
	oldRefreshToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", 24*time.Hour)
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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
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

	// Mock: Audit log (Async)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "EMAIL_VERIFIED" && req.Entity == "User"
	}), mock.Anything).Return(nil).Once()

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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	sessionID := "concurrent-session"
	numConcurrent := 20

	// Generate valid access token
	accessToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-access-secret", 15*time.Minute)
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
	taskDistributor := new(mocking.MockTaskDistributor)

	log := logrus.New()
	log.SetOutput(io.Discard)

	mockPublisher := new(mock_auth.MockNotificationPublisher)
	mockPublisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		log,
		mockPublisher,
		nil, // NIL AUTHZ
		taskDistributor,
		new(mock_auth.MockTicketManager),
	)

	user, password := createTestUser("password123")

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
	taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	authz := new(mock_auth.MockAuthzManager)

	log := logrus.New()
	log.SetOutput(io.Discard)

	mockPublisher := new(mock_auth.MockNotificationPublisher)
	mockPublisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		log,
		mockPublisher,
		authz,
		nil, // NIL TASK DISTRIBUTOR
		new(mock_auth.MockTicketManager),
	)

	user, password := createTestUser("password123")

	tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)

	tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)

	userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)
	tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	loginReq := model.LoginRequest{Username: user.Username, Password: password}
	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	// Should succeed even without audit logging
	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, refreshToken)
}

// --- Merged from auth_usecase_ticket_test.go ---
func TestAuthUseCase_GetTicket_Success(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "user-123"
	orgID := "org-456"
	sessionID := "session-789"
	role := "admin"
	username := "testuser"
	ticket := "generated-ticket-token"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusActive,
	}

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock TicketManager
	deps.ticketManager.On("CreateTicket", ctx, userID, orgID, sessionID, role, username).Return(ticket, nil)

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ticket, result)
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertExpectations(t)
}

func TestAuthUseCase_GetTicket_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "non-existent-user"
	orgID := "org-456"
	sessionID := "session-789"
	role := "admin"
	username := "testuser"

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "failed to find user")
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertNotCalled(t, "CreateTicket")
}

func TestAuthUseCase_GetTicket_UserSuspended(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "suspended-user"
	orgID := "org-456"
	sessionID := "session-789"
	role := "user"
	username := "suspended"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusSuspended,
	}

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.ErrorIs(t, err, usecase.ErrAccountSuspended)
	assert.Empty(t, result)
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertNotCalled(t, "CreateTicket")
}

func TestAuthUseCase_GetTicket_TicketManagerError(t *testing.T) {
	authService, deps := setupTest(t)
	ctx := context.Background()

	userID := "user-123"
	orgID := "org-456"
	sessionID := "session-789"
	role := "admin"
	username := "testuser"

	user := &entity.User{
		ID:       userID,
		Username: username,
		Status:   entity.UserStatusActive,
	}

	// Mock UserRepo
	deps.userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Mock TicketManager Error
	deps.ticketManager.On("CreateTicket", ctx, userID, orgID, sessionID, role, username).Return("", errors.New("redis error"))

	// Execute
	result, err := authService.GetTicket(ctx, model.UserSessionContext{
		UserID:    userID,
		OrgID:     orgID,
		SessionID: sessionID,
		Role:      role,
		Username:  username,
	})

	// Assert
	assert.Error(t, err)
	assert.Empty(t, result)
	deps.userRepo.AssertExpectations(t)
	deps.ticketManager.AssertExpectations(t)
}

// --- Merged from auth_usecase_verification_test.go ---
// setupVerificationTest creates test dependencies for verification tests
// ============================================================================
// REQUEST VERIFICATION TESTS
// ============================================================================

// ✅ POSITIVE CASE
func TestAuthUseCase_RequestVerification_Success(t *testing.T) {
	authService, deps := setupTest(t)
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

	// Mock Task Distributor (Audit Log Async)
	deps.taskDistributor.On("DistributeTaskAuditLog", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID &&
			req.Action == "VERIFICATION_EMAIL_REQUESTED" &&
			req.Entity == "User"
	}), mock.Anything).Return(nil)

	// Execute
	err := authService.RequestVerification(ctx, userID)

	// Assert
	assert.NoError(t, err)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

// ❌ NEGATIVE CASES
func TestAuthUseCase_RequestVerification_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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

	// Mock Task Distributor (Audit Log)
	deps.taskDistributor.On("DistributeTaskAuditLog", ctx, mock.Anything, mock.Anything).Return(nil)

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
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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

	// Mock Task Distributor (Audit Log Async)
	deps.taskDistributor.On("DistributeTaskAuditLog", ctx, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID &&
			req.Action == "EMAIL_VERIFIED" &&
			req.Entity == "User"
	}), mock.Anything).Return(nil)

	// Execute
	err := authService.VerifyEmail(ctx, token)

	// Assert
	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.tm.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
}

// ❌ NEGATIVE CASES
func TestAuthUseCase_VerifyEmail_InvalidToken(t *testing.T) {
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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
	authService, deps := setupTest(t)
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
	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog")
}

// 🔄 EDGE CASE
func TestAuthUseCase_VerifyEmail_AlreadyVerified(t *testing.T) {
	authService, deps := setupTest(t)
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

// // --- Merged from auth_usecase_guardian_test.go ---
// // Define specific struct for Guardian tests to be self-contained
// func createTestUser(password string) (*userEntity.User, string) {
// 	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// 	return &userEntity.User{
// 		ID:       "user-test-id",
// 		Username: "testuser",
// 		Name:     "Test User",
// 		Password: string(hashedPassword),
// 		Email:    "test@example.com",
// 		Status:   userEntity.UserStatusActive,
// 	}, password
// }

// TestAuthUseCase_Edge_UnicodeInUsername tests handling of Unicode characters in username.
func TestAuthUseCase_Edge_UnicodeInUsername(t *testing.T) {
	authService, deps := setupTest(t)
	unicodeUsername := "ユーザー名"
	user, password := createTestUser("password123")
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	}), mock.Anything).Return(nil)

	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Equal(t, unicodeUsername, loginResp.User.Username)
	deps.userRepo.AssertExpectations(t)
}

// TestAuthUseCase_Edge_LongUsername tests handling of extremely long usernames.
func TestAuthUseCase_Edge_LongUsername(t *testing.T) {
	authService, deps := setupTest(t)
	longUsername := strings.Repeat("a", 255) // Assuming 255 is DB limit or reasonably large
	user, password := createTestUser("password123")
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	}), mock.Anything).Return(nil)

	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.Equal(t, longUsername, loginResp.User.Username)
	deps.userRepo.AssertExpectations(t)
}

// TestAuthUseCase_Vulnerability_SQLInjectionInUsername tests that SQL injection payloads are treated as normal strings by the UseCase logic (Repositories should handle the safety).
func TestAuthUseCase_Vulnerability_SQLInjectionInUsername(t *testing.T) {
	authService, deps := setupTest(t)
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)

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
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
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

		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "ACCOUNT_LOCKED"
		}), mock.Anything).Return(nil).Once()

		_, _, err := authService.Login(context.Background(), loginReq)
		assert.ErrorIs(t, err, usecase.ErrAccountLocked)
	})
}

// TestAuthUseCase_ResetPassword_Edge_LongPassword tests that bcrypt failure is handled.
func TestAuthUseCase_ResetPassword_Edge_LongPassword(t *testing.T) {
	authService, deps := setupTest(t)
	token := "valid-token"
	resetToken := &authEntity.PasswordResetToken{
		Email:     "user@example.com",
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	// Password longer than 72 bytes causes bcrypt to fail
	longPassword := strings.Repeat("a", 73)

	deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.userRepo.On("FindByEmail", mock.Anything, resetToken.Email).Return(&entity.User{Email: resetToken.Email}, nil)

	err := authService.ResetPassword(context.Background(), token, longPassword)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to hash password")
	assert.Contains(t, err.Error(), "bcrypt: password length exceeds 72 bytes")
}

// TestAuthUseCase_ForgotPassword_Edge_EmailDistributorFailure tests graceful degradation.
func TestAuthUseCase_ForgotPassword_Edge_EmailDistributorFailure(t *testing.T) {
	authService, deps := setupTest(t)
	email := "user@example.com"
	user := &entity.User{ID: "user-id", Email: email}

	deps.userRepo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)

	// Mock distributor failure
	deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("queue error"))

	// Audit log should still happen (Async)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "FORGOT_PASSWORD_REQUEST"
	}), mock.Anything).Return(nil)

	err := authService.ForgotPassword(context.Background(), email)

	assert.NoError(t, err) // Should not return error
	deps.taskDistributor.AssertExpectations(t)
}
