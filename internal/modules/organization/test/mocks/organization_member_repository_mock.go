package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/stretchr/testify/mock"
)

// MockOrganizationMemberRepository is a mock implementation of OrganizationMemberRepository
type MockOrganizationMemberRepository struct {
	mock.Mock
}

func (m *MockOrganizationMemberRepository) AddMember(ctx context.Context, member *entity.OrganizationMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) FindMember(ctx context.Context, orgID, userID string) (*entity.OrganizationMember, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OrganizationMember), args.Error(1)
}

func (m *MockOrganizationMemberRepository) FindMembers(ctx context.Context, orgID string) ([]*entity.OrganizationMember, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OrganizationMember), args.Error(1)
}

func (m *MockOrganizationMemberRepository) CheckMembership(ctx context.Context, orgID, userID string) (bool, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationMemberRepository) GetMemberStatus(ctx context.Context, orgID, userID string) (string, error) {
	args := m.Called(ctx, orgID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockOrganizationMemberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID string) error {
	args := m.Called(ctx, orgID, userID, roleID)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) UpdateMemberStatus(ctx context.Context, orgID, userID, status string) error {
	args := m.Called(ctx, orgID, userID, status)
	return args.Error(0)
}
