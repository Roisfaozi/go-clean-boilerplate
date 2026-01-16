package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	querybuilder2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type userRepositoryData struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewUserRepository(db *gorm.DB, log *logrus.Logger) UserRepository {
	return &userRepositoryData{
		db:  db,
		log: log,
	}
}

func (r *userRepositoryData) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *userRepositoryData) Create(ctx context.Context, user *entity.User) error {
	if err := r.getDB(ctx).Create(user).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to create user")
		return err
	}
	return nil
}

func (r *userRepositoryData) Update(ctx context.Context, user *entity.User) error {
	if err := r.getDB(ctx).Save(user).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to update user")
		return err
	}
	return nil
}

func (r *userRepositoryData) UpdateStatus(ctx context.Context, userID, status string) error {
	if err := r.getDB(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("status", status).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to update user status")
		return err
	}
	return nil
}

func (r *userRepositoryData) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	if err := r.getDB(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find user by ID")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.getDB(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find user by email")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByToken(ctx context.Context, token string) (*entity.User, error) {
	var user entity.User
	if err := r.getDB(ctx).First(&user, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		r.log.WithContext(ctx).WithError(err).Error("failed to find user by token")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) Delete(ctx context.Context, id string) error {
	if err := r.getDB(ctx).Delete(&entity.User{}, "id = ?", id).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to delete user")
		return err
	}
	return nil
}

func (r *userRepositoryData) FindAll(ctx context.Context, filter *model.GetUserListRequest) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64
	query := r.getDB(ctx).Model(&entity.User{})

	if filter.Username != "" {
		query = query.Where("name LIKE ?", "%"+filter.Username+"%")
	}
	if filter.Email != "" {
		query = query.Where("email LIKE ?", "%"+filter.Email+"%")
	}

	// Get Total Count before pagination using a session branch
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to find all users")
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepositoryData) FindAllDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64
	query := r.getDB(ctx).Model(&entity.User{})

	// Apply Dynamic Filter
	var err error
	query, err = querybuilder2.GenerateDynamicQuery(query, &entity.User{}, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get Total Count using a session branch
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply Dynamic Sort
	query, err = querybuilder2.GenerateDynamicSort(query, &entity.User{}, filter)
	if err != nil {
		return nil, 0, err
	}

	// Apply Pagination if present
	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Limit(filter.PageSize).Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to find users dynamic")
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepositoryData) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.getDB(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) HardDeleteSoftDeletedUsers(ctx context.Context, retentionDays int) error {
	// Calculate cutoff time in milliseconds (since soft_delete uses milli)
	// retentionDays ago
	// GORM soft delete plugin stores deleted_at as unix milli. Active is 0.
	// We want records where deleted_at > 0 AND deleted_at < (now - retention)

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays).UnixMilli()

	// Unscoped() is required to see soft-deleted records.
	// We check if deleted_at is not 0 (deleted) AND deleted_at is older than cutoff.
	if err := r.getDB(ctx).Unscoped().Where("deleted_at > 0 AND deleted_at < ?", cutoffTime).Delete(&entity.User{}).Error; err != nil {
		r.log.WithContext(ctx).WithError(err).Error("failed to hard delete old users")
		return err
	}
	return nil
}
