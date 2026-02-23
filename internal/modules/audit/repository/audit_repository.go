package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type auditRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewAuditRepository(db *gorm.DB, log *logrus.Logger) usecase.AuditRepository {
	return &auditRepository{
		db:  db,
		log: log,
	}
}

func (r *auditRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *auditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	if log.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		log.ID = id.String()
	}
	return r.getDB(ctx).Create(log).Error
}

func (r *auditRepository) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AuditLog, int64, error) {
	var logs []*entity.AuditLog
	var total int64
	query := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Model(&entity.AuditLog{})

	query, err := querybuilder.GenerateDynamicQuery(query, &entity.AuditLog{}, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get Total using a session branch
	if !filter.SkipCount {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		total = -1
	}

	query, err = querybuilder.GenerateDynamicSort(query, &entity.AuditLog{}, filter)
	if err != nil {
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Limit(filter.PageSize).Offset(offset)
	}

	if err := query.Find(&logs).Error; err != nil {

		return nil, 0, err

	}

	return logs, total, nil

}

func (r *auditRepository) DeleteLogsOlderThan(ctx context.Context, cutoffTime int64) error {

	// Audit logs use created_at as unix milli

	if err := r.getDB(ctx).Where("created_at < ?", cutoffTime).Delete(&entity.AuditLog{}).Error; err != nil {

		r.log.WithContext(ctx).WithError(err).Error("Failed to prune old audit logs")

		return err

	}

	return nil

}

func (r *auditRepository) FindAllInBatches(ctx context.Context, startTime, endTime int64, batchSize int, process func([]*entity.AuditLog) error) error {
	var logs []*entity.AuditLog
	query := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Model(&entity.AuditLog{})

	if startTime > 0 {


		query = query.Where("created_at >= ?", startTime)

	}

	if endTime > 0 {

		query = query.Where("created_at <= ?", endTime)

	}

	result := query.FindInBatches(&logs, batchSize, func(tx *gorm.DB, batch int) error {

		return process(logs)

	})

	if result.Error != nil {
		r.log.WithContext(ctx).WithError(result.Error).Error("Failed to fetch audit logs in batches")
		return result.Error
	}

	return nil
}

func (r *auditRepository) CreateOutbox(ctx context.Context, outbox *entity.AuditOutbox) error {
	if outbox.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		outbox.ID = id.String()
	}
	return r.getDB(ctx).Create(outbox).Error
}

func (r *auditRepository) FindPendingOutbox(ctx context.Context, limit int) ([]*entity.AuditOutbox, error) {
	var results []*entity.AuditOutbox
	err := r.getDB(ctx).
		Where("status = ? OR (status = ? AND retry_count < ?)", entity.OutboxStatusPending, entity.OutboxStatusFailed, 5).
		Order("created_at ASC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

func (r *auditRepository) UpdateOutboxStatus(ctx context.Context, id string, status string, lastError string) error {
	updates := map[string]interface{}{
		"status":     status,
		"last_error": lastError,
	}
	if status == entity.OutboxStatusFailed {
		updates["retry_count"] = gorm.Expr("retry_count + 1")
	}
	return r.getDB(ctx).Model(&entity.AuditOutbox{}).Where("id = ?", id).Updates(updates).Error
}

func (r *auditRepository) DeleteOutbox(ctx context.Context, id string) error {
	return r.getDB(ctx).Unscoped().Delete(&entity.AuditOutbox{}, "id = ?", id).Error
}
