package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/stretchr/testify/mock"
)

// MockOrganizationMemberUseCase is a mock implementation of OrganizationMemberUseCase
type MockOrganizationMemberUseCase struct {
	mock.Mock
}

func (m *MockOrganizationMemberUseCase) InviteMember(ctx context.Context, orgID string, request *model.InviteMemberRequest) (*model.MemberResponse, error) {
	args := m.Called(ctx, orgID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberResponse), args.Error(1)
}

func (m *MockOrganizationMemberUseCase) GetMembers(ctx context.Context, orgID string) ([]model.MemberResponse, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.MemberResponse), args.Error(1)
}

func (m *MockOrganizationMemberUseCase) UpdateMember(ctx context.Context, orgID, userID string, request *model.UpdateMemberRequest) (*model.MemberResponse, error) {
	args := m.Called(ctx, orgID, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.MemberResponse), args.Error(1)
}

func (m *MockOrganizationMemberUseCase) RemoveMember(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockOrganizationMemberUseCase) AcceptInvitation(ctx context.Context, request *model.AcceptInvitationRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockOrganizationMemberUseCase) GetPresence(ctx context.Context, orgID string) ([]interface{}, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}
