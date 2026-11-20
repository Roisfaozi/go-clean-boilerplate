package test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sso"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

const (
	TestAccessSecret  = "test-access-secret"
	TestRefreshSecret = "test-refresh-secret"
	TestUserID        = "user-test-id"
	TestUsername      = "testuser"
	TestRole          = "role:user"
	TestFrontendURL   = "https://app.example.com"
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
	ticketManager   *mocking.MockTicketManager
	ssoProviders    map[string]sso.Provider
}

func setupTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	jwtManager := jwt.NewJWTManager(TestAccessSecret, TestRefreshSecret, 15*time.Minute, 24*time.Hour)

	deps := &testDependencies{
		jwtManager:      jwtManager,
		tokenRepo:       mock_auth.NewMockTokenRepository(t),
		userRepo:        mock_user.NewMockUserRepository(t),
		orgRepo:         mock_org.NewMockOrganizationRepository(t),
		tm:              mocking.NewMockWithTransactionManager(t),
		publisher:       mock_auth.NewMockNotificationPublisher(t),
		authz:           mock_auth.NewMockAuthzManager(t),
		validate:        validator.New(),
		log:             logrus.New(),
		taskDistributor: mocking.NewMockTaskDistributor(t),
		ticketManager:   mocking.NewMockTicketManager(t),
		ssoProviders:    make(map[string]sso.Provider),
	}

	deps.log.SetOutput(io.Discard)
	deps.tokenRepo.On("CountActiveSessions", mock.Anything, mock.Anything).Return(0, nil).Maybe()

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		3,
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
		deps.ssoProviders,
		TestFrontendURL,
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

func TestLogin_Failure_MaxConcurrentSessionsExceeded(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: password}

	deps.tokenRepo.ExpectedCalls = nil
	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
	deps.tokenRepo.On("CountActiveSessions", mock.Anything, user.ID).Return(3, nil)

	_, _, err := authService.Login(context.Background(), loginReq)

	assert.ErrorIs(t, err, usecase.ErrTooManySessions)
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
	deps.userRepo.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, exception.ErrNotFound)

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
		3,
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
		nil,
		TestFrontendURL,
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

	oldRefreshToken, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(oldRefreshToken))
	hashedOldRefreshToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: hashedOldRefreshToken}
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

func TestRefreshToken_ConcurrentSameToken_ReusesSingleFlow(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	oldRefreshToken, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(oldRefreshToken))
	hashedOldRefreshToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: hashedOldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil).Maybe()
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Maybe()
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil).Maybe()
	deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, "session-1").Return(nil).Maybe()
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil).Maybe()

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == "session-1"
	}), mock.Anything).Return(nil).Maybe()

	var wg sync.WaitGroup
	wg.Add(2)
	results := make(chan *model.TokenResponse, 2)
	errorsCh := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			resp, _, err := authService.RefreshToken(context.Background(), oldRefreshToken)
			results <- resp
			errorsCh <- err
		}()
	}

	wg.Wait()
	close(results)
	close(errorsCh)

	var firstResponse *model.TokenResponse
	for resp := range results {
		if firstResponse == nil {
			firstResponse = resp
			continue
		}
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, firstResponse.AccessToken, resp.AccessToken)
		assert.Equal(t, firstResponse.RefreshToken, resp.RefreshToken)
	}

	for err := range errorsCh {
		assert.NoError(t, err)
	}

	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.authz.AssertExpectations(t)
	deps.taskDistributor.AssertExpectations(t)
	assert.NotNil(t, firstResponse)
	assert.NotEmpty(t, firstResponse.AccessToken)
}

func TestRefreshToken_EnforcerError(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	oldRefreshToken, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(oldRefreshToken))
	hashedOldRefreshToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: hashedOldRefreshToken}
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

	refreshToken, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(refreshToken))
	hashedRefreshToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: hashedRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(nil, exception.ErrNotFound)

	_, _, err = authService.RefreshToken(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrNotFound))
	deps.userRepo.AssertExpectations(t)
}

func TestValidateAccessToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	token, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestAccessSecret, 15*time.Minute, "access")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, AccessToken: hashedToken}
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

	deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(&model.Auth{ID: sessionID, UserID: userID}, nil)
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

	deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(&model.Auth{ID: sessionID, UserID: userID}, nil)
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
		return payload.To == user.Email &&
			payload.Subject == "Password Reset Request" &&
			strings.Contains(payload.Body, TestFrontendURL+"/reset-password?token=")
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

	token, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: hashedToken}
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
		ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(resetToken, nil)
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
		ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(resetToken, nil)
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

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(nil, errors.New("token not found"))

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
		ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(resetToken, nil)
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
		ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(resetToken, nil)
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
		ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
	}

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(resetToken, nil)
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

	token, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(sum[:])

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: hashedToken}
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

	token, err := jwt.GenerateTestTokenWithType(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour, "refresh")
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
	deps.tokenRepo.On("CountActiveSessions", mock.Anything, user.ID).Return(0, nil)

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
		return payload.To == user.Email &&
			payload.Subject == "Verify Your Email Address" &&
			strings.Contains(payload.Body, TestFrontendURL+"/verify-email?token=")
	}), mock.Anything).Return(nil)

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "VERIFICATION_EMAIL_REQUESTED"
	})).Return(nil)

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
	assert.Equal(t, usecase.ErrAlreadyVerified, err)
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

	assert.NoError(t, err)
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
	deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(nil, exception.ErrNotFound).Once()
	deps.userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, exception.ErrNotFound)

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
	}), mock.Anything).Return(nil).Maybe()

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
	refreshToken, err := jwt.GenerateTestTokenWithType(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", 24*time.Hour, "refresh")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(refreshToken))
	hashedRefreshToken := hex.EncodeToString(sum[:])

	// Mock session valid (for ValidateRefreshToken -> validateSession)
	savedSession := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		RefreshToken: hashedRefreshToken,
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

	tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), refreshToken)

	// Revocation failure is a hard error
	assert.Error(t, err)
	assert.Nil(t, tokenResp)
	assert.Empty(t, newRefreshToken)
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
	oldAccessToken, err := jwt.GenerateTestTokenWithType(user.ID, sessionID, "role:user", user.Username, "", "test-access-secret", 15*time.Minute, "access")
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
	oldRefreshToken, err := jwt.GenerateTestTokenWithType(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", 24*time.Hour, "refresh")
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
	}), mock.Anything).Return(nil).Maybe()

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
	accessToken, err := jwt.GenerateTestTokenWithType(user.ID, sessionID, "role:user", user.Username, "", "test-access-secret", 15*time.Minute, "access")
	assert.NoError(t, err)

	sum := sha256.Sum256([]byte(accessToken))
	hashedAccessToken := hex.EncodeToString(sum[:])

	// Mock valid session
	validSession := &model.Auth{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  hashedAccessToken,
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

	tokenRepo := mock_auth.NewMockTokenRepository(t)
	userRepo := mock_user.NewMockUserRepository(t)
	orgRepo := mock_org.NewMockOrganizationRepository(t)
	tm := mocking.NewMockWithTransactionManager(t)
	taskDistributor := mocking.NewMockTaskDistributor(t)

	log := logrus.New()
	log.SetOutput(io.Discard)

	mockPublisher := mock_auth.NewMockNotificationPublisher(t)
	mockPublisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		3,
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		log,
		mockPublisher,
		nil, // NIL AUTHZ
		taskDistributor,
		mocking.NewMockTicketManager(t),
		nil,
		TestFrontendURL,
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
	tokenRepo.On("CountActiveSessions", mock.Anything, user.ID).Return(0, nil)
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

	tokenRepo := mock_auth.NewMockTokenRepository(t)
	userRepo := mock_user.NewMockUserRepository(t)
	orgRepo := mock_org.NewMockOrganizationRepository(t)
	tm := mocking.NewMockWithTransactionManager(t)
	taskDistributor := mocking.NewMockTaskDistributor(t)
	taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything).Return(nil)

	log := logrus.New()
	log.SetOutput(io.Discard)

	mockPublisher := mock_auth.NewMockNotificationPublisher(t)
	mockPublisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	authService := usecase.NewAuthUsecase(
		5,
		30*time.Minute,
		3,
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		log,
		mockPublisher,
		nil, // NIL AUTHZ
		taskDistributor,
		mocking.NewMockTicketManager(t),
		nil,
		TestFrontendURL,
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
	tokenRepo.On("CountActiveSessions", mock.Anything, user.ID).Return(0, nil)
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
	deps.userRepo.On("FindByID", ctx, userID).Return(nil, exception.ErrNotFound)

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
	deps.userRepo.On("FindByEmail", ctx, email).Return(nil, exception.ErrNotFound)

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
	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

func TestRevokeToken_NoSession_NoOp(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"

	deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(nil, nil)

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
	deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog", mock.Anything, mock.Anything)
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
		}), mock.Anything).Return(nil).Maybe()

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
		ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
	}
	// Password longer than 72 bytes causes bcrypt to fail
	longPassword := strings.Repeat("a", 73)

	deps.tokenRepo.On("FindByToken", mock.Anything, mock.Anything).Return(resetToken, nil)
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

func TestAuthUseCase_GetSSORedirectURL_Success(t *testing.T) {
	uc, deps := setupTest(t)

	ssoProvider := mock_auth.NewMockProvider(t)
	ssoProvider.On("GetLoginURL", "test-state").Return("http://sso-login-url")

	deps.ssoProviders["github"] = ssoProvider

	url, err := uc.GetSSORedirectURL(context.Background(), "github", "test-state")

	assert.NoError(t, err)
	assert.Equal(t, "http://sso-login-url", url)
	ssoProvider.AssertExpectations(t)
}

func TestAuthUseCase_GetSSORedirectURL_ProviderNotFound(t *testing.T) {
	uc, _ := setupTest(t)

	url, err := uc.GetSSORedirectURL(context.Background(), "unknown-provider", "test-state")

	assert.Error(t, err)
	assert.Equal(t, "bad request", err.Error())
	assert.Equal(t, "", url)
}

func TestAuthUseCase_HandleSSOCallback_ProviderNotFound(t *testing.T) {
	uc, _ := setupTest(t)

	res, _, err := uc.HandleSSOCallback(context.Background(), "unknown", "test-code")

	assert.Error(t, err)
	assert.Equal(t, "bad request", err.Error())
	assert.Nil(t, res)
}

func TestAuthUseCase_HandleSSOCallback_ExchangeCodeError(t *testing.T) {
	uc, deps := setupTest(t)

	ssoProvider := mock_auth.NewMockProvider(t)
	ssoProvider.On("ExchangeCode", mock.Anything, "test-code").Return(nil, errors.New("exchange error"))
	deps.ssoProviders["github"] = ssoProvider

	res, _, err := uc.HandleSSOCallback(context.Background(), "github", "test-code")

	assert.Error(t, err)
	assert.Equal(t, "unauthorized", err.Error())
	assert.Nil(t, res)
	ssoProvider.AssertExpectations(t)
}

func TestAuthUseCase_HandleSSOCallback_GetUserInfoError(t *testing.T) {
	uc, deps := setupTest(t)

	token := &oauth2.Token{AccessToken: "token"}

	ssoProvider := mock_auth.NewMockProvider(t)
	ssoProvider.On("ExchangeCode", mock.Anything, "test-code").Return(token, nil)
	ssoProvider.On("GetUserInfo", mock.Anything, token).Return(nil, errors.New("user info error"))

	deps.ssoProviders["github"] = ssoProvider

	res, _, err := uc.HandleSSOCallback(context.Background(), "github", "test-code")

	assert.Error(t, err)
	assert.Equal(t, "unauthorized", err.Error())
	assert.Nil(t, res)
	ssoProvider.AssertExpectations(t)
}

func TestAuthUseCase_HandleSSOCallback_ExistingSSOIdentity_Success(t *testing.T) {
	uc, deps := setupTest(t)

	token := &oauth2.Token{AccessToken: "token"}
	userInfo := &sso.UserInfo{Email: "test@example.com", ProviderID: "12345", Name: "Test User"}

	ssoProvider := mock_auth.NewMockProvider(t)
	ssoProvider.On("ExchangeCode", mock.Anything, "test-code").Return(token, nil)
	ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil)
	deps.ssoProviders["github"] = ssoProvider

	ssoIdentity := &entity.UserSSOIdentity{UserID: TestUserID, Provider: "github", ProviderID: "12345"}
	deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "12345").Return(ssoIdentity, nil)

	usr := &entity.User{ID: TestUserID, Email: "test@example.com", Status: entity.UserStatusActive}
	deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(usr, nil)

	deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, usr.Email).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	res, refresh, err := uc.HandleSSOCallback(context.Background(), "github", "test-code")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, refresh)
	ssoProvider.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

func TestAuthUseCase_HandleSSOCallback_ExistingEmail_LinkIdentity(t *testing.T) {
	uc, deps := setupTest(t)

	token := &oauth2.Token{AccessToken: "token"}
	userInfo := &sso.UserInfo{Email: "test@example.com", ProviderID: "12345", Name: "Test User"}

	ssoProvider := mock_auth.NewMockProvider(t)
	ssoProvider.On("ExchangeCode", mock.Anything, "test-code").Return(token, nil)
	ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil)
	deps.ssoProviders["github"] = ssoProvider

	deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "12345").Return(nil, errors.New("not found"))

	usr := &entity.User{ID: TestUserID, Email: "test@example.com", Status: entity.UserStatusActive}
	deps.userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(usr, nil)

	deps.userRepo.On("CreateSSOIdentity", mock.Anything, mock.AnythingOfType("*entity.UserSSOIdentity")).Return(nil)

	deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, usr.Email).Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything).Return(nil)

	res, refresh, err := uc.HandleSSOCallback(context.Background(), "github", "test-code")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, refresh)
	ssoProvider.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

func TestAuthUseCase_HandleSSOCallback_NewUser_AutoProvision(t *testing.T) {
	uc, deps := setupTest(t)

	token := &oauth2.Token{AccessToken: "token"}
	userInfo := &sso.UserInfo{Email: "new@example.com", ProviderID: "12345", Name: "New User"}

	ssoProvider := mock_auth.NewMockProvider(t)
	ssoProvider.On("ExchangeCode", mock.Anything, "test-code").Return(token, nil)
	ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil)
	deps.ssoProviders["github"] = ssoProvider

	deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "12345").Return(nil, errors.New("not found"))
	deps.userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))

	deps.tm.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(context.Background())
	})

	deps.userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
	deps.authz.On("AssignDefaultRole", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	deps.orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Organization"), "owner").Return(nil)

	deps.userRepo.On("CreateSSOIdentity", mock.Anything, mock.AnythingOfType("*entity.UserSSOIdentity")).Return(nil)

	deps.authz.On("GetRolesForUser", mock.Anything, mock.AnythingOfType("string"), "").Return([]string{"user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, "new@example.com").Return(nil)
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	res, refresh, err := uc.HandleSSOCallback(context.Background(), "github", "test-code")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, refresh)
	ssoProvider.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.orgRepo.AssertExpectations(t)
}

func TestRevokeToken_RepeatLogout_Idempotent(t *testing.T) {
	t.Run("second logout on missing session is no-op", func(t *testing.T) {
		authService, deps := setupTest(t)
		userID, sessionID := "user-repeat-1", "session-repeat-1"

		deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(&model.Auth{ID: sessionID, UserID: userID}, nil).Once()
		deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.UserID == userID && req.Action == "LOGOUT" && req.Entity == "Auth" && req.EntityID == sessionID
		}), mock.Anything).Return(nil).Once()

		deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(nil, nil).Once()

		assert.NoError(t, authService.RevokeToken(context.Background(), userID, sessionID))
		assert.NoError(t, authService.RevokeToken(context.Background(), userID, sessionID))

		deps.tokenRepo.AssertNumberOfCalls(t, "DeleteToken", 1)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskAuditLog", 1)
	})

	t.Run("logout when repo returns ErrNotFound is no-op", func(t *testing.T) {
		authService, deps := setupTest(t)
		userID, sessionID := "user-repeat-2", "session-repeat-2"

		deps.tokenRepo.On("GetToken", mock.Anything, userID, sessionID).Return(nil, exception.ErrNotFound).Once()

		assert.NoError(t, authService.RevokeToken(context.Background(), userID, sessionID))
		deps.tokenRepo.AssertNotCalled(t, "DeleteToken", mock.Anything, mock.Anything, mock.Anything)
		deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog", mock.Anything, mock.Anything)
	})
}

func TestVerifyEmail_RepeatWhenAlreadyVerified_AndConcurrent(t *testing.T) {
	t.Run("repeat verify on already verified user is predictable no-op", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, _ := createTestUser("password123")
		verifiedAt := time.Now().Add(-10 * time.Minute).UnixMilli()
		user.EmailVerifiedAt = &verifiedAt

		token := "stale-verification-token"
		tokenEntity := &authEntity.EmailVerificationToken{
			Email:     user.Email,
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}

		deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(tokenEntity, nil).Twice()
		deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil).Twice()
		deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil).Twice()

		assert.NoError(t, authService.VerifyEmail(context.Background(), token))
		assert.NoError(t, authService.VerifyEmail(context.Background(), token))

		deps.userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
		deps.taskDistributor.AssertNotCalled(t, "DistributeTaskAuditLog", mock.Anything, mock.Anything)
	})

	t.Run("concurrent verify with same token executes once", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, _ := createTestUser("password123")
		user.EmailVerifiedAt = nil

		token := "concurrent-verification-token"
		tokenEntity := &authEntity.EmailVerificationToken{
			Email:     user.Email,
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}

		started := make(chan struct{})
		release := make(chan struct{})
		var updateCalls atomic.Int32

		deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Run(func(args mock.Arguments) {
			close(started)
			<-release
		}).Return(tokenEntity, nil).Once()
		deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil).Once()
		deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			}).Once()
		deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).
			Run(func(args mock.Arguments) {
				updateCalls.Add(1)
			}).Return(nil).Once()
		deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "EMAIL_VERIFIED" && req.UserID == user.ID
		}), mock.Anything).Return(nil).Once()

		var wg sync.WaitGroup
		errs := make(chan error, 3)
		wg.Add(3)
		go func() {
			defer wg.Done()
			errs <- authService.VerifyEmail(context.Background(), token)
		}()
		<-started
		for i := 0; i < 2; i++ {
			go func() {
				defer wg.Done()
				errs <- authService.VerifyEmail(context.Background(), token)
			}()
		}
		time.Sleep(20 * time.Millisecond)
		close(release)
		wg.Wait()
		close(errs)

		for err := range errs {
			assert.NoError(t, err)
		}
		assert.Equal(t, int32(1), updateCalls.Load())
	})
}

func TestAuthUseCase_SSOState_OneTimeConsumption(t *testing.T) {
	t.Run("valid state consumed once and rejected on replay", func(t *testing.T) {
		uc, deps := setupTest(t)

		token := &oauth2.Token{AccessToken: "token"}
		userInfo := &sso.UserInfo{Email: "sso@example.com", ProviderID: "gh-99", Name: "SSO User"}
		state := "one-time-state-token"

		ssoProvider := mock_auth.NewMockProvider(t)
		ssoProvider.On("GetLoginURL", state).Return("https://github.com/login?state=" + state).Once()
		ssoProvider.On("ExchangeCode", mock.Anything, "code-1").Return(token, nil).Once()
		ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil).Once()
		deps.ssoProviders["github"] = ssoProvider

		ssoIdentity := &entity.UserSSOIdentity{UserID: TestUserID, Provider: "github", ProviderID: "gh-99"}
		deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "gh-99").Return(ssoIdentity, nil).Once()
		usr := &entity.User{ID: TestUserID, Email: "sso@example.com", Status: entity.UserStatusActive}
		deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(usr, nil).Once()
		deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"role:user"}, nil).Once()
		deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil).Once()
		deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, usr.Email).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		redirectURL, err := uc.GetSSORedirectURL(context.Background(), "github", state)
		assert.NoError(t, err)
		assert.Contains(t, redirectURL, state)

		// First callback with issued state succeeds and consumes state immediately.
		ctxWithState := usecase.WithSSOState(context.Background(), state)
		res, refresh, err := uc.HandleSSOCallback(ctxWithState, "github", "code-1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, refresh)

		// Second callback reusing the same state must fail with Unauthorized before ExchangeCode.
		res2, refresh2, err := uc.HandleSSOCallback(ctxWithState, "github", "code-2")
		assert.ErrorIs(t, err, exception.ErrUnauthorized)
		assert.Nil(t, res2)
		assert.Empty(t, refresh2)
		ssoProvider.AssertNotCalled(t, "ExchangeCode", mock.Anything, "code-2")
	})

	t.Run("unknown provider mismatch or empty state rejected", func(t *testing.T) {
		uc, deps := setupTest(t)

		githubProvider := mock_auth.NewMockProvider(t)
		googleProvider := mock_auth.NewMockProvider(t)
		githubProvider.On("GetLoginURL", "gh-state").Return("https://github.com/login?state=gh-state").Once()
		deps.ssoProviders["github"] = githubProvider
		deps.ssoProviders["google"] = googleProvider

		_, err := uc.GetSSORedirectURL(context.Background(), "github", "gh-state")
		assert.NoError(t, err)

		// Cross-provider state replay must fail.
		_, _, err = uc.HandleSSOCallback(usecase.WithSSOState(context.Background(), "gh-state"), "google", "code-x")
		assert.ErrorIs(t, err, exception.ErrUnauthorized)

		// Forged or empty state must fail.
		_, _, err = uc.HandleSSOCallback(usecase.WithSSOState(context.Background(), "forged-state"), "github", "code-x")
		assert.ErrorIs(t, err, exception.ErrUnauthorized)
		_, _, err = uc.HandleSSOCallback(usecase.WithSSOState(context.Background(), ""), "github", "code-x")
		assert.ErrorIs(t, err, exception.ErrUnauthorized)

		googleProvider.AssertNotCalled(t, "ExchangeCode", mock.Anything, mock.Anything)
		githubProvider.AssertNotCalled(t, "ExchangeCode", mock.Anything, mock.Anything)
	})

	t.Run("concurrent callbacks with same state allow only one winner", func(t *testing.T) {
		uc, deps := setupTest(t)

		token := &oauth2.Token{AccessToken: "token"}
		userInfo := &sso.UserInfo{Email: "race@example.com", ProviderID: "gh-race", Name: "Race User"}
		state := "concurrent-sso-state"

		ssoProvider := mock_auth.NewMockProvider(t)
		ssoProvider.On("GetLoginURL", state).Return("https://github.com/login?state=" + state).Once()
		ssoProvider.On("ExchangeCode", mock.Anything, "race-code").Return(token, nil).Once()
		ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil).Once()
		deps.ssoProviders["github"] = ssoProvider

		ssoIdentity := &entity.UserSSOIdentity{UserID: TestUserID, Provider: "github", ProviderID: "gh-race"}
		deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "gh-race").Return(ssoIdentity, nil).Once()
		usr := &entity.User{ID: TestUserID, Email: "race@example.com", Status: entity.UserStatusActive}
		deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(usr, nil).Once()
		deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"role:user"}, nil).Once()
		deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil).Once()
		deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, usr.Email).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_, err := uc.GetSSORedirectURL(context.Background(), "github", state)
		assert.NoError(t, err)

		const workers = 5
		var successes, unauthorized atomic.Int32
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				res, _, cbErr := uc.HandleSSOCallback(usecase.WithSSOState(context.Background(), state), "github", "race-code")
				if cbErr == nil && res != nil {
					successes.Add(1)
				} else if errors.Is(cbErr, exception.ErrUnauthorized) {
					unauthorized.Add(1)
				}
			}()
		}
		wg.Wait()

		assert.Equal(t, int32(1), successes.Load(), "exactly one callback may consume the SSO state")
		assert.Equal(t, int32(workers-1), unauthorized.Load(), "replayed callbacks must be rejected with ErrUnauthorized")
	})
}

func TestAuthUseCase_SSOState_Redis_OneTimeAndMultiInstance(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Skipf("miniredis unavailable in current environment: %v", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = redisClient.Close() }()

	t.Run("state stored in Redis with TTL, consumed across instances, and rejected on replay", func(t *testing.T) {
		mr.FlushAll()

		uc1, deps1 := setupTest(t)
		uc2, deps2 := setupTest(t)
		uc1.(*usecase.Service).SetRedisClient(redisClient)
		uc2.(*usecase.Service).SetRedisClient(redisClient)

		state := "redis-sso-state-1"
		token := &oauth2.Token{AccessToken: "token"}
		userInfo := &sso.UserInfo{Email: "sso-redis@example.com", ProviderID: "gh-redis-1", Name: "Redis SSO User"}

		provider1 := mock_auth.NewMockProvider(t)
		provider1.On("GetLoginURL", state).Return("https://github.com/login?state=" + state).Once()
		deps1.ssoProviders["github"] = provider1

		provider2 := mock_auth.NewMockProvider(t)
		provider2.On("ExchangeCode", mock.Anything, "code-1").Return(token, nil).Once()
		provider2.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil).Once()
		deps2.ssoProviders["github"] = provider2

		ssoIdentity := &entity.UserSSOIdentity{UserID: TestUserID, Provider: "github", ProviderID: "gh-redis-1"}
		deps2.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "gh-redis-1").Return(ssoIdentity, nil).Once()
		usr := &entity.User{ID: TestUserID, Email: "sso-redis@example.com", Status: entity.UserStatusActive}
		deps2.userRepo.On("FindByID", mock.Anything, TestUserID).Return(usr, nil).Once()
		deps2.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"role:user"}, nil).Once()
		deps2.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil).Once()
		deps2.tokenRepo.On("ResetLoginAttempts", mock.Anything, usr.Email).Return(nil).Once()
		deps2.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		redirectURL, err := uc1.GetSSORedirectURL(context.Background(), "github", state)
		assert.NoError(t, err)
		assert.Contains(t, redirectURL, state)

		redisKey := "sso:state:github:" + state
		assert.True(t, mr.Exists(redisKey))
		assert.Equal(t, 5*time.Minute, mr.TTL(redisKey))

		// Second instance consumes the state issued by the first instance.
		ctxWithState := usecase.WithSSOState(context.Background(), state)
		res, refresh, err := uc2.HandleSSOCallback(ctxWithState, "github", "code-1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, refresh)
		assert.False(t, mr.Exists(redisKey), "SSO state key must be atomically deleted upon consumption")

		// Replaying the same state on uc1 must fail before ExchangeCode.
		res2, refresh2, err := uc1.HandleSSOCallback(ctxWithState, "github", "code-replay")
		assert.ErrorIs(t, err, exception.ErrUnauthorized)
		assert.Nil(t, res2)
		assert.Empty(t, refresh2)
		provider1.AssertNotCalled(t, "ExchangeCode", mock.Anything, mock.Anything)
	})

	t.Run("expired Redis state is rejected", func(t *testing.T) {
		mr.FlushAll()

		uc, deps := setupTest(t)
		uc.(*usecase.Service).SetRedisClient(redisClient)

		state := "expired-redis-state"
		provider := mock_auth.NewMockProvider(t)
		provider.On("GetLoginURL", state).Return("https://github.com/login?state=" + state).Once()
		deps.ssoProviders["github"] = provider

		_, err := uc.GetSSORedirectURL(context.Background(), "github", state)
		assert.NoError(t, err)

		mr.FastForward(5*time.Minute + time.Second)

		res, refresh, err := uc.HandleSSOCallback(usecase.WithSSOState(context.Background(), state), "github", "code-late")
		assert.ErrorIs(t, err, exception.ErrUnauthorized)
		assert.Nil(t, res)
		assert.Empty(t, refresh)
		provider.AssertNotCalled(t, "ExchangeCode", mock.Anything, mock.Anything)
	})

	t.Run("concurrent Redis callbacks allow strictly one winner", func(t *testing.T) {
		mr.FlushAll()

		uc, deps := setupTest(t)
		uc.(*usecase.Service).SetRedisClient(redisClient)

		state := "concurrent-redis-state"
		token := &oauth2.Token{AccessToken: "token"}
		userInfo := &sso.UserInfo{Email: "race-redis@example.com", ProviderID: "gh-race-redis", Name: "Race Redis"}

		provider := mock_auth.NewMockProvider(t)
		provider.On("GetLoginURL", state).Return("https://github.com/login?state=" + state).Once()
		provider.On("ExchangeCode", mock.Anything, "race-code").Return(token, nil).Once()
		provider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil).Once()
		deps.ssoProviders["github"] = provider

		ssoIdentity := &entity.UserSSOIdentity{UserID: TestUserID, Provider: "github", ProviderID: "gh-race-redis"}
		deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "gh-race-redis").Return(ssoIdentity, nil).Once()
		usr := &entity.User{ID: TestUserID, Email: "race-redis@example.com", Status: entity.UserStatusActive}
		deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(usr, nil).Once()
		deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"role:user"}, nil).Once()
		deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil).Once()
		deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, usr.Email).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_, err := uc.GetSSORedirectURL(context.Background(), "github", state)
		assert.NoError(t, err)

		const workers = 5
		var successes, unauthorized atomic.Int32
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				res, _, cbErr := uc.HandleSSOCallback(usecase.WithSSOState(context.Background(), state), "github", "race-code")
				if cbErr == nil && res != nil {
					successes.Add(1)
				} else if errors.Is(cbErr, exception.ErrUnauthorized) {
					unauthorized.Add(1)
				}
			}()
		}
		wg.Wait()

		assert.Equal(t, int32(1), successes.Load())
		assert.Equal(t, int32(workers-1), unauthorized.Load())
	})
}

func TestLogin_ReplayWindow(t *testing.T) {
	setupLoginSuccessMocks := func(deps *testDependencies, user *entity.User, storedSessionID *string) {
		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
		deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
		deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
		deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
		deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).
			Run(func(args mock.Arguments) {
				if storedSessionID != nil {
					*storedSessionID = args.Get(1).(*model.Auth).ID
				}
			}).Return(nil)
		deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.UserID == user.ID && req.Action == "LOGIN"
		}), mock.Anything).Return(nil)
		deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()
	}

	t.Run("repeat login within replay window with same IP/UA returns same session without extra StoreToken", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, password := createTestUser("password123")
		var sessionID string
		setupLoginSuccessMocks(deps, user, &sessionID)

		req1 := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "192.0.2.10:50001", UserAgent: "Mozilla/5.0"}
		resp1, rt1, err := authService.Login(context.Background(), req1)
		assert.NoError(t, err)
		assert.NotNil(t, resp1)
		assert.NotEmpty(t, sessionID)

		deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).
			Return(&model.Auth{ID: sessionID, UserID: user.ID}, nil).Once()

		// Repeat login from same normalized IP (ephemeral port differs) and same User-Agent
		req2 := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "192.0.2.10:50002", UserAgent: "Mozilla/5.0"}
		resp2, rt2, err := authService.Login(context.Background(), req2)
		assert.NoError(t, err)
		assert.NotNil(t, resp2)
		assert.Equal(t, resp1.AccessToken, resp2.AccessToken)
		assert.Equal(t, rt1, rt2)
		assert.NotSame(t, resp1, resp2, "replayed login must return a deep copy")

		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 1)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskAuditLog", 1)
		deps.publisher.AssertNumberOfCalls(t, "PublishUserLoggedIn", 1)
	})

	t.Run("wrong password during replay window still fails with ErrInvalidCredentials", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, password := createTestUser("password123")
		var sessionID string
		setupLoginSuccessMocks(deps, user, &sessionID)

		req := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "10.0.0.1", UserAgent: "Agent-A"}
		resp1, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp1)

		deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil).Once()

		badReq := model.LoginRequest{Username: user.Username, Password: "wrong-password", IPAddress: "10.0.0.1", UserAgent: "Agent-A"}
		respBad, rtBad, err := authService.Login(context.Background(), badReq)
		assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
		assert.Nil(t, respBad)
		assert.Empty(t, rtBad)
		deps.tokenRepo.AssertNumberOfCalls(t, "IncrementLoginAttempts", 1)
	})

	t.Run("different IP or UA creates a distinct session", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, password := createTestUser("password123")
		setupLoginSuccessMocks(deps, user, nil)

		resp1, _, err := authService.Login(context.Background(), model.LoginRequest{
			Username: user.Username, Password: password, IPAddress: "10.0.0.1", UserAgent: "Agent-A",
		})
		assert.NoError(t, err)

		resp2, _, err := authService.Login(context.Background(), model.LoginRequest{
			Username: user.Username, Password: password, IPAddress: "10.0.0.1", UserAgent: "Agent-B",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, resp1.AccessToken, resp2.AccessToken)

		resp3, _, err := authService.Login(context.Background(), model.LoginRequest{
			Username: user.Username, Password: password, IPAddress: "10.0.0.2", UserAgent: "Agent-A",
		})
		assert.NoError(t, err)
		assert.NotEqual(t, resp1.AccessToken, resp3.AccessToken)
		assert.NotEqual(t, resp2.AccessToken, resp3.AccessToken)

		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 3)
	})

	t.Run("after RevokeToken immediate re-login creates a fresh session", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, password := createTestUser("password123")
		var sessionID string
		setupLoginSuccessMocks(deps, user, &sessionID)

		req := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "10.0.0.5", UserAgent: "Agent-Logout"}
		resp1, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)
		firstSessionID := sessionID

		deps.tokenRepo.On("GetToken", mock.Anything, user.ID, firstSessionID).
			Return(&model.Auth{ID: firstSessionID, UserID: user.ID}, nil).Once()
		deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, firstSessionID).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.UserID == user.ID && req.Action == "LOGOUT" && req.EntityID == firstSessionID
		}), mock.Anything).Return(nil).Once()

		assert.NoError(t, authService.RevokeToken(context.Background(), user.ID, firstSessionID))

		resp2, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)
		assert.NotEqual(t, resp1.AccessToken, resp2.AccessToken)
		assert.NotEqual(t, firstSessionID, sessionID)
		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 2)
	})

	t.Run("after RevokeAllSessions immediate re-login creates a fresh session", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, password := createTestUser("password123")
		var sessionID string
		setupLoginSuccessMocks(deps, user, &sessionID)

		req := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "10.0.0.6", UserAgent: "Agent-RevokeAll"}
		resp1, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)
		firstSessionID := sessionID

		deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.UserID == user.ID && req.Action == "REVOKE_ALL_SESSIONS"
		}), mock.Anything).Return(nil).Once()

		assert.NoError(t, authService.RevokeAllSessions(context.Background(), user.ID))

		resp2, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)
		assert.NotEqual(t, resp1.AccessToken, resp2.AccessToken)
		assert.NotEqual(t, firstSessionID, sessionID)
		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 2)
	})

	t.Run("after replay window expires re-login creates a fresh session", func(t *testing.T) {
		authService, deps := setupTest(t)
		authService.(*usecase.Service).SetLoginReplayWindow(30 * time.Millisecond)

		user, password := createTestUser("password123")
		setupLoginSuccessMocks(deps, user, nil)

		req := model.LoginRequest{Username: user.Username, Password: password, IPAddress: "10.0.0.7", UserAgent: "Agent-Expire"}
		resp1, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		resp2, _, err := authService.Login(context.Background(), req)
		assert.NoError(t, err)
		assert.NotEqual(t, resp1.AccessToken, resp2.AccessToken)
		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 2)
	})

	t.Run("register auto-login does not populate replay cache; immediate login publishes event and second login replays", func(t *testing.T) {
		authService, deps := setupTest(t)
		password := "password123"
		regReq := model.RegisterRequest{
			Username:  "reg_replay_user",
			Email:     "reg_replay@example.com",
			Password:  password,
			Name:      "Reg Replay User",
			IPAddress: "192.0.2.50:40001",
			UserAgent: "ReplayTestAgent/1.0",
		}

		var createdUser *entity.User
		deps.userRepo.On("FindByUsername", mock.Anything, regReq.Username).Return(nil, exception.ErrNotFound).Once()
		deps.userRepo.On("FindByEmail", mock.Anything, regReq.Email).Return(nil, exception.ErrNotFound).Once()
		deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Return(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})
		deps.userRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
			return u.Username == regReq.Username && u.Email == regReq.Email
		})).Run(func(args mock.Arguments) {
			createdUser = args.Get(1).(*entity.User)
		}).Return(nil).Once()
		deps.authz.On("AssignDefaultRole", mock.Anything, mock.Anything).Return(nil).Once()
		deps.orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Organization"), "owner").Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "REGISTER" && req.Entity == "User"
		}), mock.Anything).Return(nil).Once()

		deps.tokenRepo.On("IsAccountLocked", mock.Anything, regReq.Username).Return(false, time.Duration(0), nil)
		deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, regReq.Username).Return(nil)
		deps.userRepo.On("FindByUsername", mock.Anything, regReq.Username).Return(func(context.Context, string) (*entity.User, error) {
			return createdUser, nil
		}).Times(3)
		deps.authz.On("GetRolesForUser", mock.Anything, mock.Anything, "").Return([]string{TestRole}, nil)

		var sessionIDs []string
		deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).
			Run(func(args mock.Arguments) {
				sessionIDs = append(sessionIDs, args.Get(1).(*model.Auth).ID)
			}).Return(nil)
		deps.orgRepo.On("FindUserOrganizations", mock.Anything, mock.Anything).Return([]*orgEntity.Organization{
			{ID: "org-1"},
		}, nil)
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			return req.Action == "LOGIN" && req.Entity == "Auth"
		}), mock.Anything).Return(nil)
		deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, []string{"org-1"}).Return()

		// 1. Register triggers internal auto-login without populating loginReplayCache.
		regResp, regRT, err := authService.Register(context.Background(), regReq)
		assert.NoError(t, err)
		assert.NotNil(t, regResp)
		assert.NotEmpty(t, regRT)
		assert.Len(t, sessionIDs, 1)
		deps.publisher.AssertNumberOfCalls(t, "PublishUserLoggedIn", 1)

		// 2. Immediate Login from same IP/UA must NOT hit Register's auto-login cache and must publish login event.
		loginReq := model.LoginRequest{
			Username:  regReq.Username,
			Password:  password,
			IPAddress: "192.0.2.50:40002",
			UserAgent: regReq.UserAgent,
		}
		loginResp1, loginRT1, err := authService.Login(context.Background(), loginReq)
		assert.NoError(t, err)
		assert.NotNil(t, loginResp1)
		assert.NotEqual(t, regResp.AccessToken, loginResp1.AccessToken)
		assert.NotEqual(t, regRT, loginRT1)
		assert.Len(t, sessionIDs, 2)
		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 2)
		deps.publisher.AssertNumberOfCalls(t, "PublishUserLoggedIn", 2)

		// 3. Second Login immediately afterward from same IP/UA MUST hit the replay cache.
		deps.tokenRepo.On("GetToken", mock.Anything, createdUser.ID, sessionIDs[1]).
			Return(&model.Auth{ID: sessionIDs[1], UserID: createdUser.ID}, nil).Once()

		loginResp2, loginRT2, err := authService.Login(context.Background(), loginReq)
		assert.NoError(t, err)
		assert.NotNil(t, loginResp2)
		assert.Equal(t, loginResp1.AccessToken, loginResp2.AccessToken)
		assert.Equal(t, loginRT1, loginRT2)
		assert.NotSame(t, loginResp1, loginResp2, "replayed login must return a deep copy")
		deps.tokenRepo.AssertNumberOfCalls(t, "StoreToken", 2)
		deps.publisher.AssertNumberOfCalls(t, "PublishUserLoggedIn", 2)
	})
}

func TestForgotPassword_ResendCooldown(t *testing.T) {
	t.Run("in-memory cooldown suppresses repeat within window and allows retry after expiry", func(t *testing.T) {
		authService, deps := setupTest(t)
		authService.(*usecase.Service).SetResendCooldown(40 * time.Millisecond)

		user, _ := createTestUser("password123")
		deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
		deps.userRepo.On("FindByEmail", mock.Anything, " TEST@example.com ").Return(user, nil)
		deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

		// First attempt issues token and enqueues email.
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))

		// Immediate repeat (even with casing/whitespace variation) is suppressed.
		assert.NoError(t, authService.ForgotPassword(context.Background(), " TEST@example.com "))
		deps.tokenRepo.AssertNumberOfCalls(t, "Save", 1)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)

		// After cooldown expires, next attempt succeeds.
		time.Sleep(60 * time.Millisecond)
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))
		deps.tokenRepo.AssertNumberOfCalls(t, "Save", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 2)
	})

	t.Run("cooldown rolls back when tokenRepo.Save fails", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, _ := createTestUser("password123")

		deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil).Twice()
		deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).
			Return(errors.New("db error")).Once()
		deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).
			Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))

		deps.tokenRepo.AssertNumberOfCalls(t, "Save", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)
	})

	t.Run("redis-backed cooldown suppresses repeat and recovers after TTL or Save error", func(t *testing.T) {
		mr, err := miniredis.Run()
		if err != nil {
			t.Skipf("miniredis unavailable in current environment: %v", err)
		}
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		defer func() { _ = redisClient.Close() }()

		authService, deps := setupTest(t)
		authService.(*usecase.Service).SetRedisClient(redisClient)

		user, _ := createTestUser("password123")
		deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
		deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).
			Return(errors.New("transient db error")).Once()
		deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).
			Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

		// 1. Save fails -> cooldown key rolled back -> immediate retry succeeds.
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))
		deps.tokenRepo.AssertNumberOfCalls(t, "Save", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)

		// 2. Immediate repeat within 60s Redis cooldown is suppressed.
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))
		deps.tokenRepo.AssertNumberOfCalls(t, "Save", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)

		// 3. Fast-forward past 60s TTL -> retry succeeds.
		mr.FastForward(61 * time.Second)
		assert.NoError(t, authService.ForgotPassword(context.Background(), user.Email))
		deps.tokenRepo.AssertNumberOfCalls(t, "Save", 3)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 2)
	})
}

func TestRequestVerification_ResendCooldown(t *testing.T) {
	t.Run("in-memory cooldown suppresses repeat within window and allows retry after expiry", func(t *testing.T) {
		authService, deps := setupTest(t)
		authService.(*usecase.Service).SetResendCooldown(40 * time.Millisecond)

		user, _ := createTestUser("password123")
		user.EmailVerifiedAt = nil

		deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
		deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))
		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))

		deps.tokenRepo.AssertNumberOfCalls(t, "SaveVerificationToken", 1)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)

		time.Sleep(60 * time.Millisecond)
		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))
		deps.tokenRepo.AssertNumberOfCalls(t, "SaveVerificationToken", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 2)
	})

	t.Run("cooldown rolls back when SaveVerificationToken fails", func(t *testing.T) {
		authService, deps := setupTest(t)
		user, _ := createTestUser("password123")
		user.EmailVerifiedAt = nil

		deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil).Twice()
		deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).
			Return(errors.New("db error")).Once()
		deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).
			Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := authService.RequestVerification(context.Background(), user.ID)
		assert.Error(t, err)

		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))
		deps.tokenRepo.AssertNumberOfCalls(t, "SaveVerificationToken", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)
	})

	t.Run("redis-backed cooldown suppresses repeat and recovers after TTL or Save error", func(t *testing.T) {
		mr, err := miniredis.Run()
		if err != nil {
			t.Skipf("miniredis unavailable in current environment: %v", err)
		}
		defer mr.Close()

		redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		defer func() { _ = redisClient.Close() }()

		authService, deps := setupTest(t)
		authService.(*usecase.Service).SetRedisClient(redisClient)

		user, _ := createTestUser("password123")
		user.EmailVerifiedAt = nil

		deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
		deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).
			Return(errors.New("transient db error")).Once()
		deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).
			Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

		// 1. Save fails -> cooldown key rolled back -> immediate retry succeeds.
		assert.Error(t, authService.RequestVerification(context.Background(), user.ID))
		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))
		assert.True(t, mr.Exists("auth:cooldown:verify:"+user.ID))
		deps.tokenRepo.AssertNumberOfCalls(t, "SaveVerificationToken", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)

		// 2. Immediate repeat within 60s Redis cooldown is suppressed.
		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))
		deps.tokenRepo.AssertNumberOfCalls(t, "SaveVerificationToken", 2)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)

		// 3. Fast-forward past 60s TTL -> retry succeeds.
		mr.FastForward(61 * time.Second)
		assert.NoError(t, authService.RequestVerification(context.Background(), user.ID))
		deps.tokenRepo.AssertNumberOfCalls(t, "SaveVerificationToken", 3)
		deps.taskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 2)
	})
}
