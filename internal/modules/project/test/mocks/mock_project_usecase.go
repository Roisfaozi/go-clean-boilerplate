package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/stretchr/testify/mock"
)

type MockProjectUseCase struct {
	mock.Mock
}

func (m *MockProjectUseCase) CreateProject(ctx context.Context, userID string, orgID string, req model.CreateProjectRequest) (*model.ProjectResponse, error) {
	args := m.Called(ctx, userID, orgID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) GetProjects(ctx context.Context, orgID string) ([]*model.ProjectResponse, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) GetProjectByID(ctx context.Context, id string) (*model.ProjectResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) UpdateProject(ctx context.Context, id string, req model.UpdateProjectRequest) (*model.ProjectResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectResponse), args.Error(1)
}

func (m *MockProjectUseCase) DeleteProject(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
