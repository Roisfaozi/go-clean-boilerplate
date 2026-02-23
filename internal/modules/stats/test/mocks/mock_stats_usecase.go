package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/stretchr/testify/mock"
)

type MockStatsUseCase struct {
	mock.Mock
}

func (m *MockStatsUseCase) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DashboardSummary), args.Error(1)
}

func (m *MockStatsUseCase) GetDashboardActivity(ctx context.Context, days int) (*model.DashboardActivity, error) {
	args := m.Called(ctx, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.DashboardActivity), args.Error(1)
}

func (m *MockStatsUseCase) GetSystemInsights(ctx context.Context) (*model.SystemInsights, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemInsights), args.Error(1)
}
