package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type roleRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewRoleRepository(db *sqlx.DB, log *logrus.Logger) RoleRepository {
	return &roleRepository{
		db:  db,
		log: log,
	}
}

func (r *roleRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	if role.ID == "" {
		role.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if role.CreatedAt == 0 {
		role.CreatedAt = now
	}
	if role.UpdatedAt == 0 {
		role.UpdatedAt = now
	}

	query := `
		INSERT INTO roles (id, name, organization_id, description, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		role.ID, role.Name, role.OrganizationID, role.Description,
		role.CreatedAt, role.UpdatedAt,
	)
	return err
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	now := time.Now().UnixMilli()
	role.UpdatedAt = now

	query := `
		UPDATE roles
		SET description = ?, updated_at = ?
		WHERE id = ? AND deleted_at = 0
	`
	res, err := r.getDB(ctx).ExecContext(ctx, query, role.Description, role.UpdatedAt, role.ID)
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

func (r *roleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{id}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "roles.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "roles.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var role entity.Role
	err := r.getDB(ctx).GetContext(ctx, &role, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindOrganizationRoleByID(ctx context.Context, organizationID, roleID string) (*entity.Role, error) {
	whereClauses := []string{"id = ?", "organization_id = ?", "deleted_at = 0"}
	args := []any{roleID, organizationID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "roles.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var role entity.Role
	err := r.getDB(ctx).GetContext(ctx, &role, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindOrganizationRoles(ctx context.Context, organizationID string) ([]*entity.Role, error) {
	whereClauses := []string{"(organization_id = ? OR organization_id IS NULL)", "deleted_at = 0"}
	args := []any{organizationID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "roles.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var roles []*entity.Role
	err := r.getDB(ctx).SelectContext(ctx, &roles, query, args...)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	whereClauses := []string{"name = ?", "deleted_at = 0"}
	args := []any{name}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "roles.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "roles.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var role entity.Role
	err := r.getDB(ctx).GetContext(ctx, &role, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByNameInScope(ctx context.Context, name string, orgID *string) (*entity.Role, error) {
	var whereClauses []string
	var args []any

	if orgID == nil || *orgID == "" {
		whereClauses = []string{"name = ?", "organization_id IS NULL", "deleted_at = 0"}
		args = []any{name}
	} else {
		whereClauses = []string{"name = ?", "organization_id = ?", "deleted_at = 0"}
		args = []any{name, *orgID}
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var role entity.Role
	err := r.getDB(ctx).GetContext(ctx, &role, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]*entity.Role, error) {
	whereClauses := []string{"deleted_at = 0"}
	args := []any{}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "roles.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "roles.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var roles []*entity.Role
	err := r.getDB(ctx).SelectContext(ctx, &roles, query, args...)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.Role, error) {
	parsed, err := querybuilder.BuildRawQuery(&entity.Role{}, filter)
	if err != nil {
		return nil, err
	}

	whereClauses := []string{"deleted_at = 0"}
	args := []any{}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "roles.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "roles.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	if parsed.WhereSQL != "" {
		whereClauses = append(whereClauses, parsed.WhereSQL)
		args = append(args, parsed.Args...)
	}

	orderBy := "created_at DESC"
	if parsed.OrderBy != "" {
		orderBy = parsed.OrderBy
	}

	dataArgs := append([]any{}, args...)
	query := fmt.Sprintf(`
		SELECT id, name, organization_id, description, created_at, updated_at, deleted_at
		FROM roles
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, strings.Join(whereClauses, " AND "), orderBy)

	dataArgs = append(dataArgs, parsed.Limit, parsed.Offset)

	var roles []*entity.Role
	if err := r.getDB(ctx).SelectContext(ctx, &roles, query, dataArgs...); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	query := `UPDATE roles SET deleted_at = ? WHERE id = ? AND deleted_at = 0`
	res, err := r.getDB(ctx).ExecContext(ctx, query, now, id)
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

func (r *roleRepository) DeleteInOrg(ctx context.Context, orgID, roleID string) error {
	if orgID == "" || roleID == "" {
		return exception.ErrNotFound
	}

	now := time.Now().UnixMilli()
	query := `UPDATE roles SET deleted_at = ? WHERE id = ? AND organization_id = ? AND deleted_at = 0`
	res, err := r.getDB(ctx).ExecContext(ctx, query, now, roleID, orgID)
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
