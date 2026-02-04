package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/stretchr/testify/mock"
)

// MockOrganizationRepository is a mock implementation of OrganizationRepository
type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *entity.Organization, ownerRoleID string) error {
	args := m.Called(ctx, org, ownerRoleID)
	return args.Error(0)
}

func (m *MockOrganizationRepository) FindByID(ctx context.Context, id string) (*entity.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *entity.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrganizationRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationRepository) FindUserOrganizations(ctx context.Context, userID string) ([]*entity.Organization, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Organization), args.Error(1)
}
