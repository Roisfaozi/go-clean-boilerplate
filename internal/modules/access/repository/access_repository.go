package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type accessRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewAccessRepository(db *sqlx.DB, log *logrus.Logger) AccessRepository {
	return &accessRepository{
		db:  db,
		log: log,
	}
}

func (r *accessRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *accessRepository) CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
	if endpoint.ID == "" {
		endpoint.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if endpoint.CreatedAt == 0 {
		endpoint.CreatedAt = now
	}
	if endpoint.UpdatedAt == 0 {
		endpoint.UpdatedAt = now
	}

	query := `
		INSERT INTO endpoints (id, path, method, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		endpoint.ID, endpoint.Path, endpoint.Method,
		endpoint.CreatedAt, endpoint.UpdatedAt,
	)
	return err
}

func (r *accessRepository) GetEndpoints(ctx context.Context) ([]*entity.Endpoint, error) {
	query := `
		SELECT id, path, method, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE deleted_at = 0
		ORDER BY created_at DESC
	`
	var endpoints []*entity.Endpoint
	err := r.getDB(ctx).SelectContext(ctx, &endpoints, query)
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (r *accessRepository) FindEndpointsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.Endpoint, int64, error) {
	parsed, err := querybuilder.BuildRawQuery(&entity.Endpoint{}, filter)
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{"deleted_at = 0"}
	args := []any{}

	if parsed.WhereSQL != "" {
		whereClauses = append(whereClauses, parsed.WhereSQL)
		args = append(args, parsed.Args...)
	}

	var total int64
	if filter == nil || !filter.SkipCount {
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM endpoints WHERE %s`, strings.Join(whereClauses, " AND "))
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
	query := fmt.Sprintf(`
		SELECT id, path, method, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, strings.Join(whereClauses, " AND "), orderBy)

	dataArgs = append(dataArgs, parsed.Limit, parsed.Offset)

	var endpoints []*entity.Endpoint
	if err := r.getDB(ctx).SelectContext(ctx, &endpoints, query, dataArgs...); err != nil {
		return nil, 0, err
	}
	return endpoints, total, nil
}

func (r *accessRepository) GetEndpointByID(ctx context.Context, id string) (*entity.Endpoint, error) {
	query := `
		SELECT id, path, method, created_at, updated_at, deleted_at
		FROM endpoints
		WHERE id = ? AND deleted_at = 0
		LIMIT 1
	`
	var ep entity.Endpoint
	err := r.getDB(ctx).GetContext(ctx, &ep, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &ep, nil
}

func (r *accessRepository) DeleteEndpoint(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	query := `UPDATE endpoints SET deleted_at = ? WHERE id = ? AND deleted_at = 0`
	_, err := r.getDB(ctx).ExecContext(ctx, query, now, id)
	return err
}

func (r *accessRepository) CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error {
	if accessRight.ID == "" {
		accessRight.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if accessRight.CreatedAt == 0 {
		accessRight.CreatedAt = now
	}
	if accessRight.UpdatedAt == 0 {
		accessRight.UpdatedAt = now
	}

	query := `
		INSERT INTO access_rights (id, organization_id, name, description, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		accessRight.ID, accessRight.OrganizationID,
		accessRight.Name, accessRight.Description,
		accessRight.CreatedAt, accessRight.UpdatedAt,
	)
	return err
}

func (r *accessRepository) loadEndpointsForAccessRights(ctx context.Context, rights []*entity.AccessRight) error {
	if len(rights) == 0 {
		return nil
	}
	ids := make([]string, len(rights))
	rightMap := make(map[string]*entity.AccessRight, len(rights))
	for i, ar := range rights {
		ids[i] = ar.ID
		rightMap[ar.ID] = ar
		ar.Endpoints = []entity.Endpoint{}
	}

	query, args, err := sqlx.In(`
		SELECT are.access_right_id, e.id, e.path, e.method, e.created_at, e.updated_at, e.deleted_at
		FROM access_right_endpoints are
		INNER JOIN endpoints e ON are.endpoint_id = e.id
		WHERE are.access_right_id IN (?) AND e.deleted_at = 0
	`, ids)
	if err != nil {
		return err
	}

	type endpointWithARID struct {
		entity.Endpoint
		AccessRightID string `db:"access_right_id"`
	}

	var rows []endpointWithARID
	if err := r.getDB(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return err
	}

	for _, row := range rows {
		if ar, ok := rightMap[row.AccessRightID]; ok {
			ar.Endpoints = append(ar.Endpoints, row.Endpoint)
		}
	}
	return nil
}

func (r *accessRepository) GetAccessRights(ctx context.Context) ([]*entity.AccessRight, error) {
	whereClauses := []string{"deleted_at = 0"}
	args := []any{}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "access_rights.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "access_rights.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, name, description, created_at, updated_at, deleted_at
		FROM access_rights
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var accessRights []*entity.AccessRight
	if err := r.getDB(ctx).SelectContext(ctx, &accessRights, query, args...); err != nil {
		return nil, err
	}

	if err := r.loadEndpointsForAccessRights(ctx, accessRights); err != nil {
		return nil, err
	}
	return accessRights, nil
}

func (r *accessRepository) FindAccessRightsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AccessRight, int64, error) {
	parsed, err := querybuilder.BuildRawQuery(&entity.AccessRight{}, filter)
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{"deleted_at = 0"}
	args := []any{}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "access_rights.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "access_rights.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	if parsed.WhereSQL != "" {
		whereClauses = append(whereClauses, parsed.WhereSQL)
		args = append(args, parsed.Args...)
	}

	var total int64
	if filter == nil || !filter.SkipCount {
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM access_rights WHERE %s`, strings.Join(whereClauses, " AND "))
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
	query := fmt.Sprintf(`
		SELECT id, organization_id, name, description, created_at, updated_at, deleted_at
		FROM access_rights
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, strings.Join(whereClauses, " AND "), orderBy)

	dataArgs = append(dataArgs, parsed.Limit, parsed.Offset)

	var accessRights []*entity.AccessRight
	if err := r.getDB(ctx).SelectContext(ctx, &accessRights, query, dataArgs...); err != nil {
		return nil, 0, err
	}

	if err := r.loadEndpointsForAccessRights(ctx, accessRights); err != nil {
		return nil, 0, err
	}
	return accessRights, total, nil
}

func (r *accessRepository) GetAccessRightByID(ctx context.Context, id string) (*entity.AccessRight, error) {
	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{id}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "access_rights.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "access_rights.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, name, description, created_at, updated_at, deleted_at
		FROM access_rights
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var ar entity.AccessRight
	err := r.getDB(ctx).GetContext(ctx, &ar, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}

	rights := []*entity.AccessRight{&ar}
	if err := r.loadEndpointsForAccessRights(ctx, rights); err != nil {
		return nil, err
	}
	return &ar, nil
}

func (r *accessRepository) DeleteAccessRight(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{now, id}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "access_rights.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "access_rights.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		UPDATE access_rights
		SET deleted_at = ?
		WHERE %s
	`, strings.Join(whereClauses, " AND "))

	res, err := r.getDB(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (r *accessRepository) LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID string) error {
	query := `
		INSERT OR IGNORE INTO access_right_endpoints (access_right_id, endpoint_id)
		VALUES (?, ?)
	`
	// In MySQL, INSERT IGNORE is standard. In SQLite, INSERT OR IGNORE is supported.
	// MySQL driver supports INSERT IGNORE. Check error for syntax.
	_, err := r.getDB(ctx).ExecContext(ctx, query, accessRightID, endpointID)
	if err != nil {
		// Fallback for MySQL
		fallbackQuery := `
			INSERT IGNORE INTO access_right_endpoints (access_right_id, endpoint_id)
			VALUES (?, ?)
		`
		_, err = r.getDB(ctx).ExecContext(ctx, fallbackQuery, accessRightID, endpointID)
	}
	return err
}

func (r *accessRepository) UnlinkEndpointFromAccessRight(ctx context.Context, accessRightID, endpointID string) error {
	if accessRightID == "" || endpointID == "" {
		return errors.New("accessRightID and endpointID cannot be empty")
	}
	query := `
		DELETE FROM access_right_endpoints
		WHERE access_right_id = ? AND endpoint_id = ?
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query, accessRightID, endpointID)
	return err
}
