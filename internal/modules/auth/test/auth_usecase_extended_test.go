package test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	authEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_auth "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	mock_org "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	mock_user "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
)

func setupExtendedTest(t *testing.T) (usecase.AuthUseCase, *testDependencies) {
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

func TestRegister_Extended(t *testing.T) {
	authService, deps := setupExtendedTest(t)
	password := "password123"
	req := model.RegisterRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  password,
		Name:      "New User",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
	}

	t.Run("Failure_AssignDefaultRole", func(t *testing.T) {
		deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(nil, errors.New("not found")).Once()
		deps.userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, errors.New("not found")).Once()

		deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(errors.New("assign role error"))

		deps.userRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		deps.authz.On("AssignDefaultRole", mock.Anything, mock.Anything).Return(errors.New("assign role error")).Once()

		_, _, err := authService.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assign role error")
	})

	t.Run("Failure_OrgCreate", func(t *testing.T) {
		deps.userRepo.On("FindByUsername", mock.Anything, req.Username).Return(nil, errors.New("not found")).Once()
		deps.userRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, errors.New("not found")).Once()

		deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(errors.New("org create error"))

		deps.userRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		deps.authz.On("AssignDefaultRole", mock.Anything, mock.Anything).Return(nil).Once()
		deps.orgRepo.On("Create", mock.Anything, mock.Anything, "owner").Return(errors.New("org create error")).Once()

		_, _, err := authService.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "org create error")
	})
}

func TestLogin_Extended(t *testing.T) {
	authService, deps := setupExtendedTest(t)
	user, _ := createTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: "password123"}

	t.Run("Failure_IsAccountLocked_Error", func(t *testing.T) {
		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), errors.New("redis error")).Once()

		_, _, err := authService.Login(context.Background(), loginReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check account status")
	})

	t.Run("Failure_IncrementLoginAttempts_Error_But_Returns_InvalidCreds", func(t *testing.T) {
		// Mock FindByUsername success
		deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil).Once()

		// Mock Transaction execution
		deps.tm.On("WithinTransaction", mock.Anything, mock.Anything).Return(usecase.ErrInvalidCredentials).Once()

		// Mock IsAccountLocked success
		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil).Once()

		// Expect Increment to be called and return error
		deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(0, errors.New("incr error")).Once()

		// Use wrong password
		badReq := model.LoginRequest{Username: user.Username, Password: "wrongpassword"}
		_, _, err := authService.Login(context.Background(), badReq)

		assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	})

	t.Run("Failure_ResetLoginAttempts_Error_Logs_But_Succeeds", func(t *testing.T) {
		// Mock FindByUsername success
		deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil).Once()

		// Transaction succeeds even if reset fails (it's just logged)
		deps.tm.On("WithinTransaction", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(func(context.Context) error)
				_ = fn(context.Background())
			}).Return(nil).Once()

		deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil).Once()
		deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(errors.New("reset error")).Once()

		// Authz & Token Store Mocks
		deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil).Once()
		deps.tokenRepo.On("StoreToken", mock.Anything, mock.Anything).Return(nil).Once()
		deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything).Return(nil).Once()
		deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil).Once()
		deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return().Once()

		resp, _, err := authService.Login(context.Background(), loginReq)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestVerifyEmail_Extended(t *testing.T) {
	authService, deps := setupExtendedTest(t)
	token := "verify-token"

	t.Run("Failure_FindVerificationToken_Error", func(t *testing.T) {
		deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(nil, errors.New("db error")).Once()

		err := authService.VerifyEmail(context.Background(), token)
		assert.ErrorIs(t, err, usecase.ErrInvalidVerificationToken)
	})

	t.Run("Failure_DeleteVerificationToken_Error_AlreadyVerified", func(t *testing.T) {
		vt := &authEntity.EmailVerificationToken{
			Email:     "test@example.com",
			Token:     token,
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &entity.User{Email: "test@example.com"}
		verifiedTime := int64(12345)
		user.EmailVerifiedAt = &verifiedTime

		deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(vt, nil).Once()
		deps.userRepo.On("FindByEmail", mock.Anything, vt.Email).Return(user, nil).Once()
		deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, vt.Email).Return(errors.New("delete error")).Once()

		err := authService.VerifyEmail(context.Background(), token)
		assert.ErrorIs(t, err, usecase.ErrAlreadyVerified)
	})
}
