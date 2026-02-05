package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/stretchr/testify/mock"
)

type MockOrganizationUseCase struct {
	mock.Mock
}

func (m *MockOrganizationUseCase) CreateOrganization(ctx context.Context, userID string, request *model.CreateOrganizationRequest) (*model.OrganizationResponse, error) {
	args := m.Called(ctx, userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationResponse), args.Error(1)
}

func (m *MockOrganizationUseCase) GetOrganization(ctx context.Context, id string) (*model.OrganizationResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationResponse), args.Error(1)
}

func (m *MockOrganizationUseCase) GetOrganizationBySlug(ctx context.Context, slug string) (*model.OrganizationResponse, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationResponse), args.Error(1)
}

func (m *MockOrganizationUseCase) UpdateOrganization(ctx context.Context, id string, request *model.UpdateOrganizationRequest) (*model.OrganizationResponse, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrganizationResponse), args.Error(1)
}

func (m *MockOrganizationUseCase) GetUserOrganizations(ctx context.Context, userID string) (*model.UserOrganizationsResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserOrganizationsResponse), args.Error(1)
}

func (m *MockOrganizationUseCase) DeleteOrganization(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}
