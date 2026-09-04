package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type organizationMemberRepository struct {
	db *sqlx.DB
}

func NewOrganizationMemberRepository(db any) OrganizationMemberRepository {
	return &organizationMemberRepository{
		db: tx.ExtractSQLX(db),
	}
}

func (r *organizationMemberRepository) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *organizationMemberRepository) CheckMembership(ctx context.Context, orgID, userID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM organization_members
		WHERE organization_id = ? AND user_id = ? AND status = ?
	`
	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query, orgID, userID, entity.MemberStatusActive)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *organizationMemberRepository) GetMemberStatus(ctx context.Context, orgID, userID string) (string, error) {
	query := `
		SELECT status
		FROM organization_members
		WHERE organization_id = ? AND user_id = ?
		LIMIT 1
	`
	var status string
	err := r.getDB(ctx).GetContext(ctx, &status, query, orgID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return status, nil
}

func (r *organizationMemberRepository) AddMember(ctx context.Context, member *entity.OrganizationMember) error {
	if member.ID == "" {
		member.ID = uuid.NewString()
	}
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role_id, status, joined_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		member.ID, member.OrganizationID, member.UserID,
		member.RoleID, member.Status, member.JoinedAt,
	)
	return err
}

func (r *organizationMemberRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	query := `DELETE FROM organization_members WHERE organization_id = ? AND user_id = ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, orgID, userID)
	return err
}

func (r *organizationMemberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID string) error {
	query := `
		UPDATE organization_members
		SET role_id = ?
		WHERE organization_id = ? AND user_id = ?
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query, roleID, orgID, userID)
	return err
}

func (r *organizationMemberRepository) UpdateMemberStatus(ctx context.Context, orgID, userID, status string) error {
	query := `
		UPDATE organization_members
		SET status = ?
		WHERE organization_id = ? AND user_id = ?
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query, status, orgID, userID)
	return err
}

func (r *organizationMemberRepository) FindMembers(ctx context.Context, orgID string) ([]*entity.OrganizationMember, error) {
	type memberJoinRow struct {
		entity.OrganizationMember
		// User fields
		U_ID              string  `db:"u_id"`
		U_OrganizationID  *string `db:"u_organization_id"`
		U_Password        string  `db:"u_password"`
		U_Email           string  `db:"u_email"`
		U_Username        string  `db:"u_username"`
		U_Name            string  `db:"u_name"`
		U_AvatarURL       string  `db:"u_avatar_url"`
		U_Token           string  `db:"u_token"`
		U_Status          string  `db:"u_status"`
		U_EmailVerifiedAt *int64  `db:"u_email_verified_at"`
		U_CreatedAt       int64   `db:"u_created_at"`
		U_UpdatedAt       int64   `db:"u_updated_at"`
		U_DeletedAt       int64   `db:"u_deleted_at"`
		// Role fields
		R_ID             string  `db:"r_id"`
		R_Name           string  `db:"r_name"`
		R_OrganizationID *string `db:"r_organization_id"`
		R_Description    string  `db:"r_description"`
		R_CreatedAt      int64   `db:"r_created_at"`
		R_UpdatedAt      int64   `db:"r_updated_at"`
		R_DeletedAt      int64   `db:"r_deleted_at"`
	}

	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role_id, om.status, om.joined_at,
		       u.id AS u_id, u.organization_id AS u_organization_id, u.password AS u_password,
		       u.email AS u_email, u.username AS u_username, u.name AS u_name,
		       u.avatar_url AS u_avatar_url, u.token AS u_token, u.status AS u_status,
		       u.email_verified_at AS u_email_verified_at, u.created_at AS u_created_at,
		       u.updated_at AS u_updated_at, u.deleted_at AS u_deleted_at,
		       r.id AS r_id, r.name AS r_name, r.organization_id AS r_organization_id,
		       r.description AS r_description, r.created_at AS r_created_at,
		       r.updated_at AS r_updated_at, r.deleted_at AS r_deleted_at
		FROM organization_members om
		INNER JOIN users u ON om.user_id = u.id
		INNER JOIN roles r ON om.role_id = r.id
		WHERE om.organization_id = ?
		ORDER BY om.joined_at ASC
	`

	var rows []memberJoinRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, orgID)
	if err != nil {
		return nil, err
	}

	members := make([]*entity.OrganizationMember, len(rows))
	for i, row := range rows {
		mem := row.OrganizationMember
		mem.User = userEntity.User{
			ID:              row.U_ID,
			OrganizationID:  row.U_OrganizationID,
			Password:        row.U_Password,
			Email:           row.U_Email,
			Username:        row.U_Username,
			Name:            row.U_Name,
			AvatarURL:       row.U_AvatarURL,
			Token:           row.U_Token,
			Status:          row.U_Status,
			EmailVerifiedAt: row.U_EmailVerifiedAt,
			CreatedAt:       row.U_CreatedAt,
			UpdatedAt:       row.U_UpdatedAt,
		}
		mem.Role = roleEntity.Role{
			ID:             row.R_ID,
			Name:           row.R_Name,
			OrganizationID: row.R_OrganizationID,
			Description:    row.R_Description,
			CreatedAt:      row.R_CreatedAt,
			UpdatedAt:      row.R_UpdatedAt,
		}
		members[i] = &mem
	}
	return members, nil
}

func (r *organizationMemberRepository) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	query := `
		SELECT role_id
		FROM organization_members
		WHERE organization_id = ? AND user_id = ? AND status = ?
		LIMIT 1
	`
	var roleID string
	err := r.getDB(ctx).GetContext(ctx, &roleID, query, orgID, userID, entity.MemberStatusActive)
	if err != nil {
		return "", err
	}
	return roleID, nil
}

func (r *organizationMemberRepository) FindMemberForUpdate(ctx context.Context, orgID, userID string) (*entity.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role_id, status, joined_at
		FROM organization_members
		WHERE organization_id = ? AND user_id = ?
		LIMIT 1
		FOR UPDATE
	`
	var member entity.OrganizationMember
	err := r.getDB(ctx).GetContext(ctx, &member, query, orgID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

func (r *organizationMemberRepository) GetMemberRoleName(ctx context.Context, orgID, userID string) (string, error) {
	query := `
		SELECT r.name
		FROM organization_members om
		INNER JOIN roles r ON om.role_id = r.id
		WHERE om.organization_id = ? AND om.user_id = ?
		LIMIT 1
	`
	var roleName string
	err := r.getDB(ctx).GetContext(ctx, &roleName, query, orgID, userID)
	if err != nil {
		return "", err
	}
	return roleName, nil
}
