package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/stretchr/testify/mock"
)

// MockInvitationRepository is a mock implementation of InvitationRepository
type MockInvitationRepository struct {
	mock.Mock
}

func (m *MockInvitationRepository) Create(ctx context.Context, invitation *entity.InvitationToken) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockInvitationRepository) FindByToken(ctx context.Context, token string) (*entity.InvitationToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.InvitationToken), args.Error(1)
}

func (m *MockInvitationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInvitationRepository) DeleteByEmailAndOrg(ctx context.Context, email string, orgID string) error {
	args := m.Called(ctx, email, orgID)
	return args.Error(0)
}

func (m *MockInvitationRepository) CleanupExpired(ctx context.Context, now int64) error {
	args := m.Called(ctx, now)
	return args.Error(0)
}
