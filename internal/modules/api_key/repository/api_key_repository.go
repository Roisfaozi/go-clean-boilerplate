package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ApiKeyRepository interface {
	Create(ctx context.Context, apiKey *entity.ApiKey) error
	FindByHash(ctx context.Context, keyHash string) (*entity.ApiKey, error)
	FindByID(ctx context.Context, id string) (*entity.ApiKey, error)
	ListByOrg(ctx context.Context, orgID string) ([]*entity.ApiKey, error)
	Update(ctx context.Context, apiKey *entity.ApiKey) error
	Delete(ctx context.Context, id string) error
}

type apiKeyRepository struct {
	db *sqlx.DB
}

func NewApiKeyRepository(db any) ApiKeyRepository {
	return &apiKeyRepository{
		db: tx.ExtractSQLX(db),
	}
}

func (r *apiKeyRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *entity.ApiKey) error {
	if apiKey.ID == "" {
		apiKey.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if apiKey.CreatedAt == 0 {
		apiKey.CreatedAt = now
	}
	if apiKey.UpdatedAt == 0 {
		apiKey.UpdatedAt = now
	}

	query := `
		INSERT INTO api_keys (id, name, key_hash, organization_id, user_id, scopes, expires_at, last_used_at, is_active, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		apiKey.ID, apiKey.Name, apiKey.KeyHash, apiKey.OrganizationID,
		apiKey.UserID, apiKey.Scopes, apiKey.ExpiresAt, apiKey.LastUsedAt,
		apiKey.IsActive, apiKey.CreatedAt, apiKey.UpdatedAt,
	)
	return err
}

func (r *apiKeyRepository) FindByHash(ctx context.Context, keyHash string) (*entity.ApiKey, error) {
	whereClauses := []string{"key_hash = ?", "is_active = true", "deleted_at = 0"}
	args := []any{keyHash}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "api_keys.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, key_hash, organization_id, user_id, scopes, expires_at, last_used_at, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var apiKey entity.ApiKey
	err := r.getDB(ctx).GetContext(ctx, &apiKey, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) FindByID(ctx context.Context, id string) (*entity.ApiKey, error) {
	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{id}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "api_keys.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, key_hash, organization_id, user_id, scopes, expires_at, last_used_at, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var apiKey entity.ApiKey
	err := r.getDB(ctx).GetContext(ctx, &apiKey, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) ListByOrg(ctx context.Context, orgID string) ([]*entity.ApiKey, error) {
	whereClauses := []string{"organization_id = ?", "deleted_at = 0"}
	args := []any{orgID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "api_keys.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, key_hash, organization_id, user_id, scopes, expires_at, last_used_at, is_active, created_at, updated_at, deleted_at
		FROM api_keys
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var apiKeys []*entity.ApiKey
	err := r.getDB(ctx).SelectContext(ctx, &apiKeys, query, args...)
	if err != nil {
		return nil, err
	}
	return apiKeys, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *entity.ApiKey) error {
	apiKey.UpdatedAt = time.Now().UnixMilli()
	query := `
		UPDATE api_keys
		SET name = ?, scopes = ?, expires_at = ?, last_used_at = ?, is_active = ?, updated_at = ?
		WHERE id = ? AND deleted_at = 0
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		apiKey.Name, apiKey.Scopes, apiKey.ExpiresAt, apiKey.LastUsedAt,
		apiKey.IsActive, apiKey.UpdatedAt, apiKey.ID,
	)
	return err
}

func (r *apiKeyRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	query := `UPDATE api_keys SET deleted_at = ? WHERE id = ? AND deleted_at = 0`
	_, err := r.getDB(ctx).ExecContext(ctx, query, now, id)
	return err
}
