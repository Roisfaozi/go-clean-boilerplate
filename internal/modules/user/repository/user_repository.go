package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type userRepositoryData struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewUserRepository(db any, log *logrus.Logger) UserRepository {
	return &userRepositoryData{
		db:  tx.ExtractSQLX(db),
		log: log,
	}
}

func (r *userRepositoryData) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return r.db
}

func (r *userRepositoryData) Create(ctx context.Context, user *entity.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if user.CreatedAt == 0 {
		user.CreatedAt = now
	}
	if user.UpdatedAt == 0 {
		user.UpdatedAt = now
	}
	if user.Status == "" {
		user.Status = entity.UserStatusActive
	}

	query := `
		INSERT INTO users (id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		user.ID, user.OrganizationID, user.Password, user.Email,
		user.Username, user.Name, user.AvatarURL, user.Token,
		user.Status, user.EmailVerifiedAt, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to create user")
		return err
	}
	return nil
}

func (r *userRepositoryData) Update(ctx context.Context, user *entity.User) error {
	now := time.Now().UnixMilli()
	user.UpdatedAt = now

	query := `
		UPDATE users
		SET organization_id = ?, password = ?, email = ?, username = ?, name = ?,
		    avatar_url = ?, token = ?, status = ?, email_verified_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at = 0
	`
	res, err := r.getDB(ctx).ExecContext(ctx, query,
		user.OrganizationID, user.Password, user.Email, user.Username,
		user.Name, user.AvatarURL, user.Token, user.Status,
		user.EmailVerifiedAt, user.UpdatedAt, user.ID,
	)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to update user")
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *userRepositoryData) UpdateStatus(ctx context.Context, userID, status string) error {
	now := time.Now().UnixMilli()
	query := `
		UPDATE users
		SET status = ?, updated_at = ?
		WHERE id = ? AND deleted_at = 0
	`
	res, err := r.getDB(ctx).ExecContext(ctx, query, status, now, userID)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to update user status")
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *userRepositoryData) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = ? AND deleted_at = 0
		LIMIT 1
	`
	var user entity.User
	err := r.getDB(ctx).GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find user by ID")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE email = ? AND deleted_at = 0
		LIMIT 1
	`
	var user entity.User
	err := r.getDB(ctx).GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find user by email")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE username = ? AND deleted_at = 0
		LIMIT 1
	`
	var user entity.User
	err := r.getDB(ctx).GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByToken(ctx context.Context, token string) (*entity.User, error) {
	query := `
		SELECT id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE token = ? AND deleted_at = 0
		LIMIT 1
	`
	var user entity.User
	err := r.getDB(ctx).GetContext(ctx, &user, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find user by token")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) Delete(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	query := `
		UPDATE users
		SET deleted_at = ?
		WHERE id = ? AND deleted_at = 0
	`
	res, err := r.getDB(ctx).ExecContext(ctx, query, now, id)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to delete user")
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *userRepositoryData) FindAll(ctx context.Context, filter *model.GetUserListRequest) ([]*entity.User, int64, error) {
	whereClauses := []string{"users.deleted_at = 0"}
	args := []any{}

	orgID := database.GetOrganizationID(ctx)
	if orgID != "" {
		whereClauses = append(whereClauses, "users.id IN (SELECT user_id FROM organization_members WHERE organization_id = ?)")
		args = append(args, orgID)
	}

	if filter != nil {
		if filter.Username != "" {
			whereClauses = append(whereClauses, "users.name LIKE ?")
			args = append(args, "%"+filter.Username+"%")
		}
		if filter.Email != "" {
			whereClauses = append(whereClauses, "users.email LIKE ?")
			args = append(args, "%"+filter.Email+"%")
		}
	}

	whereSQL := strings.Join(whereClauses, " AND ")
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE %s`, whereSQL)

	var total int64
	if err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	limit := 10
	offset := 0
	if filter != nil {
		if filter.Limit > 0 {
			limit = filter.Limit
		}
		if filter.Page > 0 {
			offset = (filter.Page - 1) * limit
		}
	}

	dataArgs := append([]any{}, args...)
	dataArgs = append(dataArgs, limit, offset)

	dataQuery := fmt.Sprintf(`
		SELECT id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	var users []*entity.User
	if err := r.getDB(ctx).SelectContext(ctx, &users, dataQuery, dataArgs...); err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to find all users")
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepositoryData) FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.User, int64, error) {
	parsed, err := querybuilder.BuildRawQuery(&entity.User{}, filter)
	if err != nil {
		return nil, 0, err
	}

	whereClauses := []string{"users.deleted_at = 0"}
	args := []any{}

	orgID := database.GetOrganizationID(ctx)
	if orgID != "" {
		whereClauses = append(whereClauses, "users.id IN (SELECT user_id FROM organization_members WHERE organization_id = ?)")
		args = append(args, orgID)
	}

	if parsed.WhereSQL != "" {
		whereClauses = append(whereClauses, parsed.WhereSQL)
		args = append(args, parsed.Args...)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int64
	if filter == nil || !filter.SkipCount {
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM users WHERE %s`, whereSQL)
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
	dataArgs = append(dataArgs, parsed.Limit, parsed.Offset)

	query := fmt.Sprintf(`
		SELECT id, organization_id, password, email, username, name, avatar_url, token, status, email_verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereSQL, orderBy)

	var users []*entity.User
	if err := r.getDB(ctx).SelectContext(ctx, &users, query, dataArgs...); err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to find users dynamic")
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepositoryData) HardDeleteSoftDeletedUsers(ctx context.Context, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays).UnixMilli()
	query := `DELETE FROM users WHERE deleted_at > 0 AND deleted_at < ?`
	_, err := r.getDB(ctx).ExecContext(ctx, query, cutoffTime)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to hard delete old users")
		return err
	}
	return nil
}

func (r *userRepositoryData) GetByOrganization(ctx context.Context, orgID string) ([]*entity.User, error) {
	query := `
		SELECT u.id, u.organization_id, u.password, u.email, u.username, u.name, u.avatar_url, u.token, u.status, u.email_verified_at, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		WHERE u.id IN (SELECT user_id FROM organization_members WHERE organization_id = ?) AND u.deleted_at = 0
		ORDER BY u.created_at DESC
	`
	var users []*entity.User
	if err := r.getDB(ctx).SelectContext(ctx, &users, query, orgID); err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to find users by organization")
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryData) FindBySSOIdentity(ctx context.Context, provider, providerID string) (*entity.UserSSOIdentity, error) {
	query := `
		SELECT id, user_id, provider, provider_id, created_at, updated_at
		FROM user_sso_identities
		WHERE provider = ? AND provider_id = ?
		LIMIT 1
	`
	var identity entity.UserSSOIdentity
	err := r.getDB(ctx).GetContext(ctx, &identity, query, provider, providerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exception.ErrNotFound
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find sso identity")
		return nil, err
	}
	return &identity, nil
}

func (r *userRepositoryData) CreateSSOIdentity(ctx context.Context, identity *entity.UserSSOIdentity) error {
	if identity.ID == "" {
		identity.ID = uuid.NewString()
	}
	now := time.Now().UnixMilli()
	if identity.CreatedAt == 0 {
		identity.CreatedAt = now
	}
	if identity.UpdatedAt == 0 {
		identity.UpdatedAt = now
	}

	query := `
		INSERT INTO user_sso_identities (id, user_id, provider, provider_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.getDB(ctx).ExecContext(ctx, query,
		identity.ID, identity.UserID, identity.Provider, identity.ProviderID,
		identity.CreatedAt, identity.UpdatedAt,
	)
	if err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to create sso identity")
		return err
	}
	return nil
}
