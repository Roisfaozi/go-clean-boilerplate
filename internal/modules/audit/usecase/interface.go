package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type AuditRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AuditLog, int64, error)
	DeleteLogsOlderThan(ctx context.Context, cutoffTime int64) error
	FindAllInBatches(ctx context.Context, startTime, endTime int64, batchSize int, process func([]*entity.AuditLog) error) error

	// Outbox methods
	CreateOutbox(ctx context.Context, outbox *entity.AuditOutbox) error
	FindPendingOutbox(ctx context.Context, limit int) ([]*entity.AuditOutbox, error)
	UpdateOutboxStatus(ctx context.Context, id string, status string, lastError string) error
	DeleteOutbox(ctx context.Context, id string) error
}

type AuditUseCase interface {
	LogActivity(ctx context.Context, req model.CreateAuditLogRequest) error
	GetLogsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.AuditLogResponse, int64, error)
	ExportLogs(ctx context.Context, fromDate, toDate string, process func([]model.AuditLogResponse) error) error
}
