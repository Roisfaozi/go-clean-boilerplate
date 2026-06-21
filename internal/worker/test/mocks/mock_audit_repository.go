package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/stretchr/testify/mock"
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

func (_m *MockAuditRepository) CreateOutbox(ctx context.Context, outbox *entity.AuditOutbox) error {
	ret := _m.Called(ctx, outbox)
	return ret.Error(0)
}

func (_m *MockAuditRepository) FindPendingOutbox(ctx context.Context, limit int) ([]*entity.AuditOutbox, error) {
	ret := _m.Called(ctx, limit)
	return ret.Get(0).([]*entity.AuditOutbox), ret.Error(1)
}

func (_m *MockAuditRepository) UpdateOutboxStatus(ctx context.Context, id string, status string, lastError string) error {
	ret := _m.Called(ctx, id, status, lastError)
	return ret.Error(0)
}

func (_m *MockAuditRepository) DeleteOutbox(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)
	return ret.Error(0)
}
