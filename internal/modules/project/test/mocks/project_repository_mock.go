package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	ret := m.Called(ctx, project)
	return ret.Error(0)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id string) (*entity.Project, error) {
	ret := m.Called(ctx, id)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*entity.Project), ret.Error(1)
}

func (m *MockProjectRepository) GetByOrgID(ctx context.Context, orgID string) ([]*entity.Project, error) {
	ret := m.Called(ctx, orgID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]*entity.Project), ret.Error(1)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *entity.Project) error {
	ret := m.Called(ctx, project)
	return ret.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id string) error {
	ret := m.Called(ctx, id)
	return ret.Error(0)
}

func (m *MockProjectRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	ret := m.Called(ctx, userID)
	return ret.Get(0).(int64), ret.Error(1)
}
