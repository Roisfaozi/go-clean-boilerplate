package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	mock "github.com/stretchr/testify/mock"
)

type MockAuditRepository struct {
	mock.Mock
}

func (_m *MockAuditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	ret := _m.Called(ctx, log)
	return ret.Error(0)
}

func (_m *MockAuditRepository) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AuditLog, int64, error) {
	ret := _m.Called(ctx, filter)
	return ret.Get(0).([]*entity.AuditLog), ret.Get(1).(int64), ret.Error(2)
}

func (_m *MockAuditRepository) DeleteLogsOlderThan(ctx context.Context, cutoffTime int64) error {
	ret := _m.Called(ctx, cutoffTime)
	return ret.Error(0)
}

func (_m *MockAuditRepository) FindAllInBatches(ctx context.Context, startTime, endTime int64, batchSize int, process func([]*entity.AuditLog) error) error {
	ret := _m.Called(ctx, startTime, endTime, batchSize, process)
	return ret.Error(0)
}
