package test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	workerMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	txMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	wsMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type authTestDeps struct {
	TokenRepo       *authMocks.MockTokenRepository
	UserRepo        *userMocks.MockUserRepository
	TM              *txMocks.MockWithTransactionManager
	Enforcer        *permissionMocks.IEnforcer
	AuditUC         *auditMocks.MockAuditUseCase
	TaskDistributor *workerMocks.MockTaskDistributor
	WSManager       *wsMocks.MockWebSocketManager
	JWTManager      *jwt.JWTManager
}

func setupAuthTest() (*authTestDeps, usecase.AuthUseCase) {
	deps := &authTestDeps{
		TokenRepo:       new(authMocks.MockTokenRepository),
		UserRepo:        new(userMocks.MockUserRepository),
		TM:              new(txMocks.MockWithTransactionManager),
		Enforcer:        new(permissionMocks.IEnforcer),
		AuditUC:         new(auditMocks.MockAuditUseCase),
		TaskDistributor: new(workerMocks.MockTaskDistributor),
		WSManager:       new(wsMocks.MockWebSocketManager),
		JWTManager:      jwt.NewJWTManager("secret", "secret", 15*time.Minute, 720*time.Hour),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewAuthUsecase(
		deps.JWTManager,
		deps.TokenRepo,
		deps.UserRepo,
		deps.TM,
		log,
		deps.WSManager,
		deps.Enforcer,
		deps.AuditUC,
		deps.TaskDistributor,
	)

	return deps, uc
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	deps, uc := setupAuthTest()

	hashedPassword, _ := pkg.HashPassword("password")
	user := &userEntity.User{
		ID:       "user-123",
		Username: "testuser",
		Password: hashedPassword,
		Email:    "test@example.com",
		Name:     "Test User",
	}

	req := model.LoginRequest{
		Username: "testuser",
		Password: "password",
	}

	// Mocking Transaction Manager
	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		_ = fn(context.Background())
	})

	deps.UserRepo.On("FindByUsername", mock.Anything, "testuser").Return(user, nil)
	deps.Enforcer.On("GetRolesForUser", "user-123").Return([]string{"admin"}, nil)
	deps.TokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)
	deps.WSManager.On("BroadcastToChannel", "global_notifications", mock.Anything).Return(nil)

	res, refreshToken, err := uc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, "testuser", res.User.Username)
	assert.Equal(t, "admin", res.User.Role)

	deps.UserRepo.AssertExpectations(t)
	deps.TokenRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

func TestAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	deps, uc := setupAuthTest()

	req := model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	// Mocking Transaction Manager to fail
	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(usecase.ErrInvalidCredentials).Run(func(args mock.Arguments) {
		// In a real scenario, this closure would call UserRepo, but since we return error from TM wrapper or inside,
		// if we want to simulate the logic inside, we need to mock UserRepo call.
		// However, s.tm.WithinTransaction implementation calls the function.
		// Let's simulate the function execution failing.
		fn := args.Get(1).(func(context.Context) error)

		// Setup expectations for inside the transaction
		deps.UserRepo.On("FindByUsername", mock.Anything, "testuser").Return(nil, errors.New("not found"))

		_ = fn(context.Background())
	})

	res, _, err := uc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidCredentials, err)
	assert.Nil(t, res)
}

func TestAuthUseCase_ForgotPassword_Success(t *testing.T) {
	deps, uc := setupAuthTest()
	email := "test@example.com"

	user := &userEntity.User{ID: "u1", Email: email}

	deps.UserRepo.On("FindByEmail", mock.Anything, email).Return(user, nil)
	deps.TokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
	deps.TaskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything).Return(nil)
	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := uc.ForgotPassword(context.Background(), email)

	assert.NoError(t, err)
	deps.UserRepo.AssertExpectations(t)
	deps.TokenRepo.AssertExpectations(t)
	deps.TaskDistributor.AssertExpectations(t)
}

func TestAuthUseCase_ForgotPassword_UserNotFound(t *testing.T) {
	deps, uc := setupAuthTest()
	email := "unknown@example.com"

	deps.UserRepo.On("FindByEmail", mock.Anything, email).Return(nil, errors.New("not found"))

	err := uc.ForgotPassword(context.Background(), email)

	// Should not return error for security reasons
	assert.NoError(t, err)
	deps.UserRepo.AssertExpectations(t)
	// TokenRepo should NOT be called
	deps.TokenRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestAuthUseCase_ResetPassword_Success(t *testing.T) {
	deps, uc := setupAuthTest()
	token := "valid-token"
	newPassword := "NewPass123!"

	resetToken := &authEntity.PasswordResetToken{
		Email:     "test@example.com",
		Token:     token,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	user := &userEntity.User{ID: "u1", Email: "test@example.com", Password: "oldhash"}

	deps.TokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.UserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)

		deps.UserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *userEntity.User) bool {
			return u.ID == "u1" && u.Password != "oldhash"
		})).Return(nil)

		deps.TokenRepo.On("DeleteByEmail", mock.Anything, "test@example.com").Return(nil)

		_ = fn(context.Background())
	})

	deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

	err := uc.ResetPassword(context.Background(), token, newPassword)

	assert.NoError(t, err)
	deps.TokenRepo.AssertExpectations(t)
	deps.UserRepo.AssertExpectations(t)
}

func TestAuthUseCase_ResetPassword_TokenExpired(t *testing.T) {
	deps, uc := setupAuthTest()
	token := "expired-token"

	resetToken := &authEntity.PasswordResetToken{
		Email:     "test@example.com",
		Token:     token,
		ExpiresAt: time.Now().Add(-10 * time.Minute),
	}

	deps.TokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
	deps.TokenRepo.On("DeleteByEmail", mock.Anything, "test@example.com").Return(nil)

	err := uc.ResetPassword(context.Background(), token, "pass")

	assert.Equal(t, usecase.ErrInvalidResetToken, err)
}

// Vulnerability Test: Ensure Token Generation fails gracefully if store fails
func TestAuthUseCase_Login_StoreTokenFail(t *testing.T) {
	deps, uc := setupAuthTest()

	hashedPassword, _ := pkg.HashPassword("password")
	user := &userEntity.User{ID: "u1", Username: "user", Password: hashedPassword}
	req := model.LoginRequest{Username: "user", Password: "password"}

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		deps.UserRepo.On("FindByUsername", mock.Anything, "user").Return(user, nil)
		_ = fn(context.Background())
	})

	deps.Enforcer.On("GetRolesForUser", "u1").Return([]string{"user"}, nil)
	deps.TokenRepo.On("StoreToken", mock.Anything, mock.Anything).Return(errors.New("redis error"))

	res, _, err := uc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store session")
	assert.Nil(t, res)
}
