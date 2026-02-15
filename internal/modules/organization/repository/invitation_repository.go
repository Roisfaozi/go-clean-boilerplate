package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"gorm.io/gorm"
)

// InvitationRepository defines the interface for invitation token data access
type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.InvitationToken) error
	FindByToken(ctx context.Context, token string) (*entity.InvitationToken, error)
	Delete(ctx context.Context, id string) error
	DeleteByEmailAndOrg(ctx context.Context, email, orgID string) error
	CleanupExpired(ctx context.Context, now int64) error
}

type invitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository creates a new InvitationRepository
func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *invitationRepository) Create(ctx context.Context, invitation *entity.InvitationToken) error {
	return r.getDB(ctx).Create(invitation).Error
}

func (r *invitationRepository) FindByToken(ctx context.Context, token string) (*entity.InvitationToken, error) {
	var invitation entity.InvitationToken
	err := r.getDB(ctx).Where("token = ?", token).First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) Delete(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&entity.InvitationToken{}, "id = ?", id).Error
}

func (r *invitationRepository) DeleteByEmailAndOrg(ctx context.Context, email, orgID string) error {
	return r.getDB(ctx).Where("email = ? AND organization_id = ?", email, orgID).Delete(&entity.InvitationToken{}).Error
}

func (r *invitationRepository) CleanupExpired(ctx context.Context, now int64) error {
	return r.getDB(ctx).Where("expires_at < ?", now).Delete(&entity.InvitationToken{}).Error
}
