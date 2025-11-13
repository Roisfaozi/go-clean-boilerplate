package mocks

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/stretchr/testify/mock"
)

// MockAuthUseCase is a mock implementation of the AuthUseCase interface
type MockAuthUseCase struct {
	mock.Mock
}

// GenerateAccessToken mocks the GenerateAccessToken method
func (m *MockAuthUseCase) GenerateAccessToken(user *entity.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

// GenerateRefreshToken mocks the GenerateRefreshToken method
func (m *MockAuthUseCase) GenerateRefreshToken(user *entity.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

// ValidateAccessToken mocks the ValidateAccessToken method
func (m *MockAuthUseCase) ValidateAccessToken(token string) (*usecase.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.Claims), args.Error(1)
}

// ValidateRefreshToken mocks the ValidateRefreshToken method
func (m *MockAuthUseCase) ValidateRefreshToken(token string) (*usecase.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.Claims), args.Error(1)
}

// RevokeToken mocks the RevokeToken method
func (m *MockAuthUseCase) RevokeToken(ctx context.Context, userID, sessionID string) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

// Login mocks the Login method
func (m *MockAuthUseCase) Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, string, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.LoginResponse), args.String(1), args.Error(2)
}

// RefreshToken mocks the RefreshToken method
func (m *MockAuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenResponse, string, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.TokenResponse), args.String(1), args.Error(2)
}

// Verify mocks the Verify method
func (m *MockAuthUseCase) Verify(ctx context.Context, userID string, sessionID string) (*model.Auth, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Auth), args.Error(1)
}

// GetUserSessions mocks the GetUserSessions method
func (m *MockAuthUseCase) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Auth), args.Error(1)
}

// RevokeAllSessions mocks the RevokeAllSessions method
func (m *MockAuthUseCase) RevokeAllSessions(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
