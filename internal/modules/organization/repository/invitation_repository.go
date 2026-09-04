package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.InvitationToken) error
	FindByToken(ctx context.Context, token string) (*entity.InvitationToken, error)
	Delete(ctx context.Context, id string) error
	DeleteByEmailAndOrg(ctx context.Context, email, orgID string) error
	CleanupExpired(ctx context.Context, now int64) error
}

type invitationRepository struct {
	db *sqlx.DB
}

func NewInvitationRepository(db any) InvitationRepository {
	return &invitationRepository{
		db: tx.ExtractSQLX(db),
	}
}

func (r *invitationRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *invitationRepository) Create(ctx context.Context, invitation *entity.InvitationToken) error {
	if invitation.ID == "" {
		invitation.ID = uuid.NewString()
	}
	query := `
		INSERT INTO invitation_tokens (id, organization_id, email, token, role_id, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		invitation.ID, invitation.OrganizationID, invitation.Email,
		invitation.Token, invitation.RoleID, invitation.ExpiresAt, invitation.CreatedAt,
	)
	return err
}

func (r *invitationRepository) FindByToken(ctx context.Context, token string) (*entity.InvitationToken, error) {
	query := `
		SELECT id, organization_id, email, token, role_id, expires_at, created_at
		FROM invitation_tokens
		WHERE token = ?
		LIMIT 1
	`
	var invitation entity.InvitationToken
	err := r.getDB(ctx).GetContext(ctx, &invitation, query, token)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM invitation_tokens WHERE id = ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, id)
	return err
}

func (r *invitationRepository) DeleteByEmailAndOrg(ctx context.Context, email, orgID string) error {
	query := `DELETE FROM invitation_tokens WHERE email = ? AND organization_id = ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, email, orgID)
	return err
}

func (r *invitationRepository) CleanupExpired(ctx context.Context, now int64) error {
	query := `DELETE FROM invitation_tokens WHERE expires_at < ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, now)
	return err
}
