package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type webhookRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewWebhookRepository(db any, log *logrus.Logger) WebhookRepository {
	return &webhookRepository{
		db:  tx.ExtractSQLX(db),
		log: log,
	}
}

func (r *webhookRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *webhookRepository) Create(ctx context.Context, webhook *entity.Webhook) error {
	if webhook.ID == "" {
		webhook.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if webhook.CreatedAt == 0 {
		webhook.CreatedAt = now
	}
	if webhook.UpdatedAt == 0 {
		webhook.UpdatedAt = now
	}

	query := `
		INSERT INTO webhooks (id, name, organization_id, url, events, secret, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		webhook.ID, webhook.Name, webhook.OrganizationID, webhook.URL,
		webhook.Events, webhook.Secret, webhook.IsActive, webhook.CreatedAt, webhook.UpdatedAt,
	)
	return err
}

func (r *webhookRepository) Update(ctx context.Context, webhook *entity.Webhook) error {
	now := time.Now().UnixMilli()
	webhook.UpdatedAt = now

	query := `
		UPDATE webhooks
		SET name = ?, url = ?, events = ?, secret = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND organization_id = ? AND deleted_at = 0
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		webhook.Name, webhook.URL, webhook.Events, webhook.Secret,
		webhook.IsActive, webhook.UpdatedAt, webhook.ID, webhook.OrganizationID,
	)
	return err
}

func (r *webhookRepository) Delete(ctx context.Context, id string, organizationID string) error {
	now := time.Now().UnixMilli()
	whereClauses := []string{"id = ?", "organization_id = ?", "deleted_at = 0"}
	args := []any{now, id, organizationID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "webhooks.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		UPDATE webhooks
		SET deleted_at = ?
		WHERE %s
	`, strings.Join(whereClauses, " AND "))

	_, err := r.getDB(ctx).ExecContext(ctx, query, args...)
	return err
}

func (r *webhookRepository) FindByID(ctx context.Context, id string, organizationID string) (*entity.Webhook, error) {
	whereClauses := []string{"id = ?", "organization_id = ?", "deleted_at = 0"}
	args := []any{id, organizationID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "webhooks.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, url, events, secret, is_active, created_at, updated_at
		FROM webhooks
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var webhook entity.Webhook
	err := r.getDB(ctx).GetContext(ctx, &webhook, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &webhook, nil
}

func (r *webhookRepository) FindByOrganizationID(ctx context.Context, organizationID string) ([]entity.Webhook, error) {
	whereClauses := []string{"organization_id = ?", "deleted_at = 0"}
	args := []any{organizationID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "webhooks.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, url, events, secret, is_active, created_at, updated_at
		FROM webhooks
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var webhooks []entity.Webhook
	err := r.getDB(ctx).SelectContext(ctx, &webhooks, query, args...)
	if err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (r *webhookRepository) FindByEvent(ctx context.Context, organizationID string, event string) ([]entity.Webhook, error) {
	whereClauses := []string{
		"organization_id = ?",
		"is_active = true",
		"deleted_at = 0",
		"JSON_CONTAINS(events, JSON_QUOTE(?))",
	}
	args := []any{organizationID, event}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "webhooks.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, url, events, secret, is_active, created_at, updated_at
		FROM webhooks
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var webhooks []entity.Webhook
	err := r.getDB(ctx).SelectContext(ctx, &webhooks, query, args...)
	if err != nil {
		return nil, err
	}
	return webhooks, nil
}

func (r *webhookRepository) CreateLog(ctx context.Context, log *entity.WebhookLog) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	if log.CreatedAt == 0 {
		log.CreatedAt = time.Now().UnixMilli()
	}

	query := `
		INSERT INTO webhook_logs (id, webhook_id, event_type, payload, response_status_code, response_body, execution_time, error_message, retry_count, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		log.ID, log.WebhookID, log.EventType, log.Payload,
		log.ResponseStatusCode, log.ResponseBody, log.ExecutionTime,
		log.ErrorMessage, log.RetryCount, log.CreatedAt,
	)
	return err
}

func (r *webhookRepository) FindLogsByWebhookID(ctx context.Context, webhookID string, limit int, offset int) ([]entity.WebhookLog, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, response_status_code, response_body, execution_time, error_message, retry_count, created_at
		FROM webhook_logs
		WHERE webhook_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	var logs []entity.WebhookLog
	err := r.getDB(ctx).SelectContext(ctx, &logs, query, webhookID, limit, offset)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
