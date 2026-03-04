package repository

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"gorm.io/gorm"
)

// organizationMemberRepository implements OrganizationMemberRepository interface.
type organizationMemberRepository struct {
	db *gorm.DB
}

// NewOrganizationMemberRepository creates a new instance of OrganizationMemberRepository.
func NewOrganizationMemberRepository(db *gorm.DB) OrganizationMemberRepository {
	return &organizationMemberRepository{db: db}
}

// CheckMembership verifies if a user is an active member of an organization.
func (r *organizationMemberRepository) CheckMembership(ctx context.Context, orgID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ? AND status = ?", orgID, userID, entity.MemberStatusActive).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMemberStatus returns the membership status of a user in an organization.
func (r *organizationMemberRepository) GetMemberStatus(ctx context.Context, orgID, userID string) (string, error) {
	var member entity.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // Not a member
		}
		return "", err
	}
	return member.Status, nil
}

// AddMember adds a user to an organization with a specific role.
func (r *organizationMemberRepository) AddMember(ctx context.Context, member *entity.OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember removes a user from an organization.
func (r *organizationMemberRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	return r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Delete(&entity.OrganizationMember{}).Error
}

// UpdateMemberRole updates a member's role in an organization.
func (r *organizationMemberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID string) error {
	return r.db.WithContext(ctx).
		Model(&entity.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("role_id", roleID).Error
}

// UpdateMemberStatus updates a member's status (active, suspended, banned).
func (r *organizationMemberRepository) UpdateMemberStatus(ctx context.Context, orgID, userID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entity.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		Update("status", status).Error
}

// FindMembers finds all members of an organization.
func (r *organizationMemberRepository) FindMembers(ctx context.Context, orgID string) ([]*entity.OrganizationMember, error) {
	var members []*entity.OrganizationMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Where("organization_id = ?", orgID).
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetMemberRole returns the role of a user in an organization.
func (r *organizationMemberRepository) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	var member entity.OrganizationMember
	err := r.db.WithContext(ctx).
		Select("role_id").
		Where("organization_id = ? AND user_id = ? AND status = ?", orgID, userID, entity.MemberStatusActive).
		First(&member).Error

	if err != nil {
		return "", err
	}
	return member.RoleID, nil
}
