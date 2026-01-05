package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	assert.Empty(t, loginResp.User.Role) // Role should be empty
	deps.enforcer.AssertExpectations(t)
}
