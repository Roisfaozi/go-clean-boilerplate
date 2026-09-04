package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id string) (*entity.Project, error)
	GetByOrgID(ctx context.Context, orgID string) ([]*entity.Project, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

type projectRepository struct {
	db *sqlx.DB
}

func NewProjectRepository(db *sqlx.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	if project.ID == "" {
		project.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if project.CreatedAt == 0 {
		project.CreatedAt = now
	}
	if project.UpdatedAt == 0 {
		project.UpdatedAt = now
	}
	if project.Status == "" {
		project.Status = "active"
	}

	query := `
		INSERT INTO projects (id, organization_id, user_id, name, domain, status, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		project.ID, project.OrganizationID, project.UserID,
		project.Name, project.Domain, project.Status,
		project.CreatedAt, project.UpdatedAt,
	)
	return err
}

func (r *projectRepository) GetByID(ctx context.Context, id string) (*entity.Project, error) {
	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{id}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "projects.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "projects.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, user_id, name, domain, status, created_at, updated_at, deleted_at
		FROM projects
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var p entity.Project
	err := r.getDB(ctx).GetContext(ctx, &p, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *projectRepository) GetByOrgID(ctx context.Context, orgID string) ([]*entity.Project, error) {
	whereClauses := []string{"organization_id = ?", "deleted_at = 0"}
	args := []any{orgID}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "projects.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, user_id, name, domain, status, created_at, updated_at, deleted_at
		FROM projects
		WHERE %s
		ORDER BY created_at DESC
	`, strings.Join(whereClauses, " AND "))

	var projects []*entity.Project
	err := r.getDB(ctx).SelectContext(ctx, &projects, query, args...)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, project *entity.Project) error {
	now := time.Now().UnixMilli()
	project.UpdatedAt = now

	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{project.Name, project.Domain, project.Status, project.UpdatedAt, project.ID}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "projects.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "projects.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		UPDATE projects
		SET name = ?, domain = ?, status = ?, updated_at = ?
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

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	whereClauses := []string{"id = ?", "deleted_at = 0"}
	args := []any{now, id}

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, "projects.organization_id")
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	visClause, visArgs := database.SQLOrganizationVisibilityClause(ctx, "projects.organization_id")
	if visClause != "" {
		whereClauses = append(whereClauses, visClause)
		args = append(args, visArgs...)
	}

	query := fmt.Sprintf(`
		UPDATE projects
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

func (r *projectRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM projects
		WHERE user_id = ? AND deleted_at = 0
	`
	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
