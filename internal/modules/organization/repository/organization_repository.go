package repository

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// organizationRepository implements OrganizationRepository interface.
type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new instance of OrganizationRepository.
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

// Create creates a new organization with the owner as the first member atomically.
func (r *organizationRepository) Create(ctx context.Context, org *entity.Organization, ownerRoleID string) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		// Create organization
		if err := tx.Create(org).Error; err != nil {
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
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		return nil
	})
}

// FindByID finds an organization by its ID.
func (r *organizationRepository) FindByID(ctx context.Context, id string) (*entity.Organization, error) {
	var org entity.Organization
	err := r.getDB(ctx).
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
	err := r.getDB(ctx).
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
	err := r.getDB(ctx).
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
	err := r.getDB(ctx).
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
	return r.getDB(ctx).
		Save(org).Error
}

// Delete soft-deletes an organization.
func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).
		Where("id = ?", id).
		Delete(&entity.Organization{}).Error
}
