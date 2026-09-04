package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type auditRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewAuditRepository(db any, log *logrus.Logger) usecase.AuditRepository {
	return &auditRepository{
		db:  tx.ExtractSQLX(db),
		log: log,
	}
}

func (r *auditRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *auditRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	if log.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		log.ID = id.String()
	}
	if log.CreatedAt == 0 {
		log.CreatedAt = time.Now().UnixMilli()
	}

	query := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		log.ID, log.OrganizationID, log.UserID, log.Action,
		log.Entity, log.EntityID, log.OldValues, log.NewValues,
		log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	return err
}

func (r *auditRepository) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AuditLog, int64, error) {
	parsed, err := querybuilder.BuildRawQuery(&entity.AuditLog{}, filter)
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{"audit_logs.deleted_at = 0"}
	args := []any{}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "audit_logs.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "audit_logs.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	if parsed.WhereSQL != "" {
		whereClauses = append(whereClauses, parsed.WhereSQL)
		args = append(args, parsed.Args...)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int64
	if filter == nil || !filter.SkipCount {
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs WHERE %s`, whereSQL)
		if err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...); err != nil {
			return nil, 0, err
		}
	} else {
		total = -1
	}

	orderBy := "created_at DESC"
	if parsed.OrderBy != "" {
		orderBy = parsed.OrderBy
	}

	dataArgs := append([]any{}, args...)
	dataArgs = append(dataArgs, parsed.Limit, parsed.Offset)

	query := fmt.Sprintf(`
		SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at
		FROM audit_logs
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereSQL, orderBy)

	var logs []*entity.AuditLog
	if err := r.getDB(ctx).SelectContext(ctx, &logs, query, dataArgs...); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *auditRepository) DeleteLogsOlderThan(ctx context.Context, cutoffTime int64) error {
	query := `DELETE FROM audit_logs WHERE created_at < ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, cutoffTime)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("Failed to prune old audit logs")
		return err
	}
	return nil
}

func (r *auditRepository) FindAllInBatches(ctx context.Context, startTime, endTime int64, batchSize int, process func([]*entity.AuditLog) error) error {
	whereClauses := []string{"audit_logs.deleted_at = 0"}
	args := []any{}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "audit_logs.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "audit_logs.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	if startTime > 0 {
		whereClauses = append(whereClauses, "created_at >= ?")
		args = append(args, startTime)
	}

	if endTime > 0 {
		whereClauses = append(whereClauses, "created_at <= ?")
		args = append(args, endTime)
	}

	whereSQL := strings.Join(whereClauses, " AND ")
	offset := 0

	for {
		batchArgs := append([]any{}, args...)
		batchArgs = append(batchArgs, batchSize, offset)

		query := fmt.Sprintf(`
			SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, created_at, deleted_at
			FROM audit_logs
			WHERE %s
			ORDER BY created_at ASC
			LIMIT ? OFFSET ?
		`, whereSQL)

		var logs []*entity.AuditLog
		if err := r.getDB(ctx).SelectContext(ctx, &logs, query, batchArgs...); err != nil {
			r.log.WithContext(ctx).WithError(err).Error("Failed to fetch audit logs in batches")
			return err
		}

		if len(logs) == 0 {
			break
		}

		if err := process(logs); err != nil {
			return err
		}

		if len(logs) < batchSize {
			break
		}
		offset += len(logs)
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
	now := time.Now().UnixMilli()
	if outbox.CreatedAt == 0 {
		outbox.CreatedAt = now
	}
	if outbox.UpdatedAt == 0 {
		outbox.UpdatedAt = now
	}
	if outbox.Status == "" {
		outbox.Status = entity.OutboxStatusPending
	}

	query := `
		INSERT INTO audit_outbox (id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, status, retry_count, last_error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		outbox.ID, outbox.OrganizationID, outbox.UserID,
		outbox.Action, outbox.Entity, outbox.EntityID,
		outbox.OldValues, outbox.NewValues, outbox.IPAddress,
		outbox.UserAgent, outbox.Status, outbox.RetryCount,
		outbox.LastError, outbox.CreatedAt, outbox.UpdatedAt,
	)
	return err
}

func (r *auditRepository) FindPendingOutbox(ctx context.Context, limit int) ([]*entity.AuditOutbox, error) {
	query := `
		SELECT id, organization_id, user_id, action, entity, entity_id, old_values, new_values, ip_address, user_agent, status, retry_count, last_error, created_at, updated_at
		FROM audit_outbox
		WHERE status = ? OR (status = ? AND retry_count < 5)
		ORDER BY created_at ASC
		LIMIT ?
	`
	var results []*entity.AuditOutbox
	err := r.getDB(ctx).SelectContext(ctx, &results, query, entity.OutboxStatusPending, entity.OutboxStatusFailed, limit)
	return results, err
}

func (r *auditRepository) UpdateOutboxStatus(ctx context.Context, id string, status string, lastError string) error {
	now := time.Now().UnixMilli()
	if status == entity.OutboxStatusFailed {
		query := `
			UPDATE audit_outbox
			SET status = ?, last_error = ?, retry_count = retry_count + 1, updated_at = ?
			WHERE id = ?
		`
		_, err := r.getDB(ctx).ExecContext(ctx, query, status, lastError, now, id)
		return err
	}

	query := `
		UPDATE audit_outbox
		SET status = ?, last_error = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query, status, lastError, now, id)
	return err
}

func (r *auditRepository) DeleteOutbox(ctx context.Context, id string) error {
	query := `DELETE FROM audit_outbox WHERE id = ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, id)
	return err
}
