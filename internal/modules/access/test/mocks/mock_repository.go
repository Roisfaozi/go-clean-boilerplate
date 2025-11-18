package mocks

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"github.com/stretchr/testify/mock"
)

type MockAccessRepository struct {
	mock.Mock
}

func (m *MockAccessRepository) CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error {
	args := m.Called(ctx, accessRight)
	return args.Error(0)
}

func (m *MockAccessRepository) FindAccessRightByName(ctx context.Context, name string) (*entity.AccessRight, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessRight), args.Error(1)
}

func (m *MockAccessRepository) GetAllAccessRights(ctx context.Context) ([]entity.AccessRight, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessRight), args.Error(1)
}

func (m *MockAccessRepository) CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
	args := m.Called(ctx, endpoint)
	return args.Error(0)
}

func (m *MockAccessRepository) GetEndpointByPathAndMethod(ctx context.Context, path, method string) (*entity.Endpoint, error) {
	args := m.Called(ctx, path, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Endpoint), args.Error(1)
}

func (m *MockAccessRepository) LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID uint) error {
	args := m.Called(ctx, accessRightID, endpointID)
	return args.Error(0)
}

func (m *MockAccessRepository) UnlinkEndpointFromAccessRight(ctx context.Context, accessRightID, endpointID uint) error {
	args := m.Called(ctx, accessRightID, endpointID)
	return args.Error(0)
}

func (m *MockAccessRepository) GetEndpointsForAccessRight(ctx context.Context, accessRightID uint) ([]entity.Endpoint, error) {
	args := m.Called(ctx, accessRightID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Endpoint), args.Error(1)
}
