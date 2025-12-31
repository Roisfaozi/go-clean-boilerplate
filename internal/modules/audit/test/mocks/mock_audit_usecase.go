package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/stretchr/testify/mock"
)

type MockAuditUseCase struct {
	mock.Mock
}

func (m *MockAuditUseCase) LogActivity(ctx context.Context, req model.CreateAuditLogRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuditUseCase) GetLogsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.AuditLogResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AuditLogResponse), args.Error(1)
}
