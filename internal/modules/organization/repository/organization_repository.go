package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type organizationRepository struct {
	db    *sqlx.DB
	redis *redis.Client
}

func NewOrganizationRepository(db any, redisClients ...*redis.Client) OrganizationRepository {
	var redisClient *redis.Client
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}
	return &organizationRepository{
		db:    tx.ExtractSQLX(db),
		redis: redisClient,
	}
}

func (r *organizationRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *organizationRepository) invalidateOrganizationStatusCache(ctx context.Context, orgID string) {
	if r.redis == nil {
		return
	}
	cacheKey := fmt.Sprintf("nexusos:org_status:%s", orgID)
	_ = r.redis.Del(ctx, cacheKey).Err()
}

func (r *organizationRepository) Create(ctx context.Context, org *entity.Organization, ownerRoleID string) error {
	createOps := func(dbtx tx.DBTX) error {
		if org.ID == "" {
			org.ID = uuid.NewString()
		}
		now := time.Now().UnixMilli()
		if org.CreatedAt == 0 {
			org.CreatedAt = now
		}
		if org.UpdatedAt == 0 {
			org.UpdatedAt = now
		}
		if org.Status == "" {
			org.Status = entity.OrgStatusActive
		}

		queryOrg := `
			INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 0)
		`
		_, err := dbtx.ExecContext(ctx, queryOrg,
			org.ID, org.Name, org.Slug, org.OwnerID, org.Status,
			org.CreatedAt, org.UpdatedAt,
		)
		if err != nil {
			return err
		}

		memberID := uuid.NewString()
		queryMember := `
			INSERT INTO organization_members (id, organization_id, user_id, role_id, status, joined_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		_, err = dbtx.ExecContext(ctx, queryMember,
			memberID, org.ID, org.OwnerID, ownerRoleID,
			entity.MemberStatusActive, now,
		)
		return err
	}

	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return createOps(dbtx)
	}

	// Not in a transaction context, manage one
	txx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = txx.Rollback()
	}()

	if err := createOps(txx); err != nil {
		return err
	}

	return txx.Commit()
}

func (r *organizationRepository) FindByID(ctx context.Context, id string) (*entity.Organization, error) {
	where := "id = ? AND deleted_at = 0"
	if database.CanAccessDeletedOrganizations(ctx) {
		where = "id = ?"
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, owner_id, status, created_at, updated_at, deleted_at
		FROM organizations
		WHERE %s
		LIMIT 1
	`, where)

	var org entity.Organization
	err := r.getDB(ctx).GetContext(ctx, &org, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) FindBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	where := "slug = ? AND deleted_at = 0"
	if database.CanAccessDeletedOrganizations(ctx) {
		where = "slug = ?"
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, owner_id, status, created_at, updated_at, deleted_at
		FROM organizations
		WHERE %s
		LIMIT 1
	`, where)

	var org entity.Organization
	err := r.getDB(ctx).GetContext(ctx, &org, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *organizationRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	query := `SELECT COUNT(*) FROM organizations WHERE slug = ?`
	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query, slug)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *organizationRepository) FindUserOrganizations(ctx context.Context, userID string) ([]*entity.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.owner_id, o.status, o.created_at, o.updated_at, o.deleted_at
		FROM organizations o
		INNER JOIN organization_members om ON om.organization_id = o.id
		WHERE om.user_id = ? AND om.status = ? AND o.deleted_at = 0
		ORDER BY o.created_at DESC
	`
	var orgs []*entity.Organization
	err := r.getDB(ctx).SelectContext(ctx, &orgs, query, userID, entity.MemberStatusActive)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *organizationRepository) Update(ctx context.Context, org *entity.Organization) error {
	now := time.Now().UnixMilli()
	org.UpdatedAt = now

	query := `
		UPDATE organizations
		SET name = ?, slug = ?, owner_id = ?, status = ?, updated_at = ?
		WHERE id = ? AND deleted_at = 0
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		org.Name, org.Slug, org.OwnerID, org.Status, org.UpdatedAt, org.ID,
	)
	return err
}

func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	query := `UPDATE organizations SET deleted_at = ? WHERE id = ? AND deleted_at = 0`
	_, err := r.getDB(ctx).ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}
	r.invalidateOrganizationStatusCache(ctx, id)
	return nil
}

func (r *organizationRepository) Restore(ctx context.Context, id string) error {
	query := `UPDATE organizations SET deleted_at = 0 WHERE id = ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	r.invalidateOrganizationStatusCache(ctx, id)
	return nil
}

func (r *organizationRepository) HardDelete(ctx context.Context, id string) error {
	query := `DELETE FROM organizations WHERE id = ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	r.invalidateOrganizationStatusCache(ctx, id)
	return nil
}
