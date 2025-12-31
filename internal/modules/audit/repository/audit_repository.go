package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type auditRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewAuditRepository creates a new instance of AuditRepository.
func NewAuditRepository(db *gorm.DB, log *logrus.Logger) usecase.AuditRepository {
	return &auditRepository{
		db:  db,
		log: log,
	}
}

// Create inserts a new audit log record.
func (r *auditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	if log.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		log.ID = id.String()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

// FindAllDynamic retrieves audit logs based on dynamic filters.
func (r *auditRepository) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AuditLog, error) {
	var logs []*entity.AuditLog
	query := r.db.WithContext(ctx)

	query, err := querybuilder.GenerateDynamicQuery(query, &entity.AuditLog{}, filter)
	if err != nil {
		return nil, err
	}

	query, err = querybuilder.GenerateDynamicSort(query, &entity.AuditLog{}, filter)
	if err != nil {
		return nil, err
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}