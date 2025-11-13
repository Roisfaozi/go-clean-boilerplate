package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/mocking"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_auth "github.com/Roisfaozi/casbin-db/internal/modules/auth/test/mocks"
	mock_user "github.com/Roisfaozi/casbin-db/internal/modules/user/test/mocks"
	mock_utils "github.com/Roisfaozi/casbin-db/internal/utils/test/mocks"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Mock dependencies
type testDependencies struct {
	config    *mock_auth.MockConfig
	tokenRepo *mock_auth.MockTokenRepository
	userRepo  *mock_user.MockUserRepository
	tm        *mocking.MockTransactionManager
	wsManager *mock_utils.MockWebSocketManager
	validate  *validator.Validate
	log       *logrus.Logger
}

// setupTest initializes all dependencies for the test
func setupTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
	deps := &testDependencies{
		config:    new(mock_auth.MockConfig),
		tokenRepo: new(mock_auth.MockTokenRepository),
		userRepo:  new(mock_user.MockUserRepository),
		tm:        new(mocking.MockTransactionManager),
		wsManager: new(mock_utils.MockWebSocketManager),
		validate:  validator.New(),
		log:       logrus.New(),
	}

	deps.log.SetOutput(&mock_utils.NoOpWriter{})

	//
	// MOVED HERE: Set up expectations for methods called inside NewService
	//
	deps.config.On("GetAccessTokenDuration").Return(15 * time.Minute)
	deps.config.On("GetRefreshTokenDuration").Return(24 * time.Hour)

	authService := usecase.NewService(
		deps.config,
		deps.tokenRepo,
		deps.userRepo,
		deps.validate,
		deps.tm,
		deps.log,
		deps.wsManager,
	)

	return authService, deps
}

func createTestUser(password string) (*entity.User, string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &entity.User{
		ID:       "user-test-id",
		Name:     "testuser",
		Password: string(hashedPassword),
	}, password
}

// --- LOGIN TESTS ---

func TestLogin_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Name, Password: password}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Name).Return(user, nil)

	// Expectations for methods called inside Login() can stay here
	deps.config.On("GetAccessTokenSecret").Return("access-secret")
	deps.config.On("GetRefreshTokenSecret").Return("refresh-secret")

	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.wsManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return()

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.AccessToken)
	assert.Equal(t, "Bearer", loginResp.TokenType)
	assert.NotEmpty(t, refreshToken)
	deps.tm.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
	deps.tokenRepo.AssertExpectations(t)
	deps.wsManager.AssertExpectations(t)
}

func TestLogin_Failure_InvalidRequest(t *testing.T) {
	authService, _ := setupTest(t)
	loginReq := model.LoginRequest{Username: "", Password: "123"} // Invalid username

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
}

func TestLogin_Failure_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	loginReq := model.LoginRequest{Username: "nonexistent", Password: "password123"}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
}

func TestLogin_Failure_InvalidPassword(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Name, Password: "wrong-password"}

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Name).Return(user, nil)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.userRepo.AssertExpectations(t)
}

func TestLogin_Failure_StoreTokenError(t *testing.T) {
	authService, deps := setupTest(t)
	user, password := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Name, Password: password}
	storeErr := errors.New("redis is down")

	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			fn(context.Background())
		}).Return(nil)
	deps.userRepo.On("FindByUsername", mock.Anything, user.Name).Return(user, nil)
	deps.config.On("GetAccessTokenSecret").Return("access-secret")
	deps.config.On("GetRefreshTokenSecret").Return("refresh-secret")
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(storeErr)

	loginResp, refreshToken, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store session")
	assert.Nil(t, loginResp)
	assert.Empty(t, refreshToken)
	deps.tokenRepo.AssertExpectations(t)
}

// --- REFRESH TOKEN TESTS ---

func TestRefreshToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.config.On("GetAccessTokenSecret").Return("access-secret")
	deps.config.On("GetRefreshTokenSecret").Return("refresh-secret")

	oldRefreshToken, err := usecase.GenerateTestToken(user.ID, "session-1", deps.config.GetRefreshTokenSecret(), deps.config.GetRefreshTokenDuration())
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, "session-1").Return(nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

	tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), oldRefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, tokenResp)
	assert.NotEmpty(t, tokenResp.AccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)
	deps.tokenRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

func TestRefreshToken_Failure_InvalidToken(t *testing.T) {
	authService, deps := setupTest(t)
	deps.config.On("GetRefreshTokenSecret").Return("refresh-secret")

	_, _, err := authService.RefreshToken(context.Background(), "this.is.an.invalid.token")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrInvalidToken))
}

func TestRefreshToken_Failure_UserNotFound(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.config.On("GetRefreshTokenSecret").Return("refresh-secret")

	refreshToken, err := usecase.GenerateTestToken(user.ID, "session-1", deps.config.GetRefreshTokenSecret(), deps.config.GetRefreshTokenDuration())
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: refreshToken}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
	deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(nil, gorm.ErrRecordNotFound)

	_, _, err = authService.RefreshToken(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	deps.userRepo.AssertExpectations(t)
}

// --- VALIDATE TOKEN TESTS ---

func TestValidateAccessToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.config.On("GetAccessTokenSecret").Return("access-secret")

	token, err := usecase.GenerateTestToken(user.ID, "session-1", deps.config.GetAccessTokenSecret(), deps.config.GetAccessTokenDuration())
	assert.NoError(t, err)

	session := &model.Auth{ID: "session-1", UserID: user.ID, AccessToken: token}
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)

	claims, err := authService.ValidateAccessToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "session-1", claims.SessionID)
	deps.tokenRepo.AssertExpectations(t)
}

func TestValidateAccessToken_Failure_Expired(t *testing.T) {
	authService, deps := setupTest(t)
	deps.config.On("GetAccessTokenSecret").Return("access-secret")

	expiredToken, err := usecase.GenerateTestToken("user-id", "session-1", "access-secret", -1*time.Hour)
	assert.NoError(t, err)

	claims, err := authService.ValidateAccessToken(expiredToken)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrExpiredToken))
	assert.Nil(t, claims)
}

func TestValidateAccessToken_Failure_TokenRevoked(t *testing.T) {
	authService, deps := setupTest(t)
	user, _ := createTestUser("password123")

	deps.config.On("GetAccessTokenSecret").Return("access-secret")

	token, err := usecase.GenerateTestToken(user.ID, "session-1", deps.config.GetAccessTokenSecret(), deps.config.GetAccessTokenDuration())
	assert.NoError(t, err)

	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(nil, nil)

	claims, err := authService.ValidateAccessToken(token)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, usecase.ErrTokenRevoked))
	assert.Nil(t, claims)
	deps.tokenRepo.AssertExpectations(t)
}

// --- REVOKE TOKEN TESTS ---

func TestRevokeToken_Success(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"

	deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(nil)

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
}

func TestRevokeToken_Failure(t *testing.T) {
	authService, deps := setupTest(t)
	userID, sessionID := "user-1", "session-1"
	revokeErr := errors.New("failed to delete")

	deps.tokenRepo.On("DeleteToken", mock.Anything, userID, sessionID).Return(revokeErr)

	err := authService.RevokeToken(context.Background(), userID, sessionID)

	assert.Error(t, err)
	assert.Equal(t, revokeErr, err)
	deps.tokenRepo.AssertExpectations(t)
}
