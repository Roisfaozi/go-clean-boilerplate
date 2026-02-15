package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockOrganizationReader is a mock implementation of IOrganizationReader
type MockOrganizationReader struct {
	mock.Mock
}

func (m *MockOrganizationReader) ValidateMembership(ctx context.Context, orgID, userID string) (bool, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationReader) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	args := m.Called(ctx, orgID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockOrganizationReader) InvalidateMembershipCache(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockOrganizationReader) InvalidateOrganizationCache(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}
