package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	txpkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// organizationRepository implements OrganizationRepository interface.
type organizationRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewOrganizationRepository creates a new instance of OrganizationRepository.
// An optional Redis client can be passed to enable cache invalidation for organization status.
func NewOrganizationRepository(db *gorm.DB, redisClients ...*redis.Client) OrganizationRepository {
	var redisClient *redis.Client
	if len(redisClients) > 0 {
		redisClient = redisClients[0]
	}
	return &organizationRepository{db: db, redis: redisClient}
}

// Create creates a new organization with the owner as the first member atomically.
func (r *organizationRepository) Create(ctx context.Context, org *entity.Organization, ownerRoleID string) error {
	createWithDB := func(db *gorm.DB) error {
		// Create organization
		if err := db.Create(org).Error; err != nil {
			return err
		}

		// Create owner as first member
		member := &entity.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: org.ID,
			UserID:         org.OwnerID,
			RoleID:         ownerRoleID,
			Status:         entity.MemberStatusActive,
		}
		if err := db.Create(member).Error; err != nil {
			return err
		}

		return nil
	}

	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return createWithDB(txDB.WithContext(ctx))
	}

	return r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		return createWithDB(db)
	})
}

func (r *organizationRepository) query(ctx context.Context) *gorm.DB {
	query := r.getDB(ctx).WithContext(ctx)
	if database.CanAccessDeletedOrganizations(ctx) {
		return query.Unscoped()
	}
	return query.Unscoped().Where("(organizations.deleted_at = 0 OR organizations.deleted_at IS NULL)")
}

func (r *organizationRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := txpkg.DBFromContext(ctx); ok {
		return txDB
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

// FindByID finds an organization by its ID.
func (r *organizationRepository) FindByID(ctx context.Context, id string) (*entity.Organization, error) {
	var org entity.Organization
	err := r.query(ctx).
		Where("id = ?", id).
		First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// FindBySlug finds an organization by its unique slug.
func (r *organizationRepository) FindBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	var org entity.Organization
	err := r.query(ctx).
		Where("slug = ?", slug).
		First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// SlugExists checks if a slug is already taken.
func (r *organizationRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Organization{}).
		Where("slug = ?", slug).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindUserOrganizations finds all organizations a user is a member of.
func (r *organizationRepository) FindUserOrganizations(ctx context.Context, userID string) ([]*entity.Organization, error) {
	var orgs []*entity.Organization
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN organization_members ON organization_members.organization_id = organizations.id").
		Where("organization_members.user_id = ?", userID).
		Where("organization_members.status = ?", entity.MemberStatusActive).
		Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

// Update updates an organization's details.
func (r *organizationRepository) Update(ctx context.Context, org *entity.Organization) error {
	return r.getDB(ctx).WithContext(ctx).
		Save(org).Error
}

// Delete soft-deletes an organization.
func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	if err := r.getDB(ctx).WithContext(ctx).
		Where("id = ?", id).
		Delete(&entity.Organization{}).Error; err != nil {
		return err
	}

	r.invalidateOrganizationStatusCache(ctx, id)
	return nil
}

// Restore clears the soft-delete marker for an organization.
func (r *organizationRepository) Restore(ctx context.Context, id string) error {
	if err := r.getDB(ctx).WithContext(ctx).
		Unscoped().
		Model(&entity.Organization{}).
		Where("id = ?", id).
		Update("deleted_at", 0).Error; err != nil {
		return err
	}

	r.invalidateOrganizationStatusCache(ctx, id)
	return nil
}

// HardDelete permanently removes an organization.
func (r *organizationRepository) HardDelete(ctx context.Context, id string) error {
	if err := r.getDB(ctx).WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&entity.Organization{}).Error; err != nil {
		return err
	}

	r.invalidateOrganizationStatusCache(ctx, id)
	return nil
}
