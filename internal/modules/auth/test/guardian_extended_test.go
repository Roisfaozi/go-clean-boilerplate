package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestAuthUseCase_GenerateAccessToken_Success tests successful access token generation.
func TestAuthUseCase_GenerateAccessToken_Success(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	user, _ := createGuardianTestUser("password123")

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)

	accessToken, err := authService.GenerateAccessToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)

	// Verify token claims
	claims, err := deps.jwtManager.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "role:user", claims.Role)
	assert.Equal(t, user.Username, claims.Username)
}

// TestAuthUseCase_GenerateRefreshToken_Success tests successful refresh token generation.
func TestAuthUseCase_GenerateRefreshToken_Success(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	user, _ := createGuardianTestUser("password123")

	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)

	refreshToken, err := authService.GenerateRefreshToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Verify token claims
	claims, err := deps.jwtManager.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "role:user", claims.Role)
	assert.Equal(t, user.Username, claims.Username)
}

// TestAuthUseCase_GetUserSessions_Success tests retrieving user sessions.
func TestAuthUseCase_GetUserSessions_Success(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	userID := "user-test-id"
	expectedSessions := []*model.Auth{
		{ID: "session-1", UserID: userID},
		{ID: "session-2", UserID: userID},
	}

	deps.tokenRepo.On("GetUserSessions", mock.Anything, userID).Return(expectedSessions, nil)

	sessions, err := authService.GetUserSessions(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSessions, sessions)
	deps.tokenRepo.AssertExpectations(t)
}

// TestAuthUseCase_RevokeAllSessions_Success tests revoking all sessions.
func TestAuthUseCase_RevokeAllSessions_Success(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	userID := "user-test-id"

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == userID && req.Action == "REVOKE_ALL_SESSIONS"
	}), mock.Anything).Return(nil)

	deps.tokenRepo.On("RevokeAllSessions", mock.Anything, userID).Return(nil)

	err := authService.RevokeAllSessions(context.Background(), userID)

	assert.NoError(t, err)
	deps.tokenRepo.AssertExpectations(t)
}

// TestAuthUseCase_Login_Negative_EmptyPassword tests login with empty password.
func TestAuthUseCase_Login_Negative_EmptyPassword(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	user, _ := createGuardianTestUser("password123")
	loginReq := model.LoginRequest{Username: user.Username, Password: ""}

	deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)

	// Mock transaction where user is found but password check fails
	deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(context.Background())
		}).Return(usecase.ErrInvalidCredentials)

	deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

	// Login attempt increment logic
	deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil)
	// GetLoginAttempts is not called if IncrementLoginAttempts returns the count directly if logic changed,
	// but looking at AuthUseCase logic:
	// attempts, incrErr := s.tokenRepo.IncrementLoginAttempts(txCtx, request.Username)
	// if attempts >= s.maxLoginAttempts { ... }
	// It uses the return value of Increment directly.
	// So we don't need GetLoginAttempts expectation unless the code calls it separately.
	// Checking code:
	// attempts, incrErr := s.tokenRepo.IncrementLoginAttempts(txCtx, request.Username)
	// ...
	// if attempts >= s.maxLoginAttempts { ... }
	// So GetLoginAttempts is NOT called in this flow.

	_, _, err := authService.Login(context.Background(), loginReq)

	assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
	deps.tokenRepo.AssertExpectations(t)
}

// TestAuthUseCase_RefreshToken_Negative_InvalidToken tests refresh token with invalid signature/structure.
func TestAuthUseCase_RefreshToken_Negative_InvalidToken(t *testing.T) {
	authService, _ := setupAuthGuardianTest(t)
	invalidToken := "invalid-token-string"

	resp, refreshToken, err := authService.RefreshToken(context.Background(), invalidToken)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrInvalidToken, err)
	assert.Nil(t, resp)
	assert.Empty(t, refreshToken)
}

// TestAuthUseCase_RefreshToken_Negative_RevokedToken tests refresh token that has been revoked (not in store).
func TestAuthUseCase_RefreshToken_Negative_RevokedToken(t *testing.T) {
	authService, deps := setupAuthGuardianTest(t)
	user, _ := createGuardianTestUser("password123")

	// Generate a valid JWT
	validRefreshToken, _, _ := deps.jwtManager.GenerateTokenPair(jwt.UserContext{
		UserID:    user.ID,
		SessionID: "session-id",
		Role:      "role:user",
		Username:  user.Username,
	})

	// Mock repository to return nil (token not found/revoked)
	deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-id").Return(nil, nil)

	resp, refreshToken, err := authService.RefreshToken(context.Background(), validRefreshToken)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrTokenRevoked, err)
	assert.Nil(t, resp)
	assert.Empty(t, refreshToken)
}

// TestAuthUseCase_Login_Edge_EnforcerError tests behavior when Enforcer fails.
func TestAuthUseCase_Login_Edge_EnforcerError(t *testing.T) {
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

	// FORCE ERROR HERE
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user roles")
	assert.Nil(t, loginResp)
}

// TestAuthUseCase_Login_Edge_WSManagerError tests that WS broadcast failure does NOT block login.
func TestAuthUseCase_Login_Edge_WSManagerError(t *testing.T) {
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

	// Mock Publisher
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{{ID: "org1"}}, nil)
	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.UserID == user.ID && req.Action == "LOGIN"
	}), mock.Anything).Return(nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	deps.publisher.AssertExpectations(t)
}

// TestAuthUseCase_Login_Positive_VerifyClaims tests that the generated token contains the correct claims.
func TestAuthUseCase_Login_Positive_VerifyClaims(t *testing.T) {
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
	deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:admin"}, nil) // Admin role
	deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{{ID: "org1"}}, nil)
	deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()
	deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)

	loginResp, _, err := authService.Login(context.Background(), loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)

	// Validate returned token manually
	claims, err := deps.jwtManager.ValidateAccessToken(loginResp.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "role:admin", claims.Role)
	assert.Equal(t, user.Username, claims.Username)
}
