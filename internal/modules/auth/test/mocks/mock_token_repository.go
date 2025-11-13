package mocks

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/stretchr/testify/mock"
)

// MockTokenRepository is a mock implementation of the TokenRepository interface
type MockTokenRepository struct {
	mock.Mock
}

// StoreToken mocks the StoreToken method
func (m *MockTokenRepository) StoreToken(ctx context.Context, session *model.Auth) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

// GetToken mocks the GetToken method
func (m *MockTokenRepository) GetToken(ctx context.Context, userID, sessionID string) (*model.Auth, error) {
	args := m.Called(ctx, userID, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Auth), args.Error(1)
}

// DeleteToken mocks the DeleteToken method
func (m *MockTokenRepository) DeleteToken(ctx context.Context, userID, sessionID string) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

// GetUserSessions mocks the GetUserSessions method
func (m *MockTokenRepository) GetUserSessions(ctx context.Context, userID string) ([]*model.Auth, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Auth), args.Error(1)
}

// RevokeAllSessions mocks the RevokeAllSessions method
func (m *MockTokenRepository) RevokeAllSessions(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
