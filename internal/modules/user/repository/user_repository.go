package repository

import (
	"context"
	"errors"

	"github.com/Roisfaozi/casbin-db/internal/modules/user/entity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type userRepositoryData struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewUserRepository creates a new instance of UserRepository.
//
// Parameters:
// - db: The GORM database connection.
// - log: The logger instance.
//
// Returns:
// - A pointer to the newly created UserRepository instance.
func NewUserRepository(db *gorm.DB, log *logrus.Logger) UserRepository {
	return &userRepositoryData{
		db:  db,
		log: log,
	}
}

func (r *userRepositoryData) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.log.WithError(err).Error("failed to create user")
		return err
	}
	return nil
}

func (r *userRepositoryData) Update(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.log.WithError(err).Error("failed to update user")
		return err
	}
	return nil
}

func (r *userRepositoryData) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		r.log.WithError(err).Error("failed to find user by ID")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		r.log.WithError(err).Error("failed to find user by email")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) FindByToken(ctx context.Context, token string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		r.log.WithError(err).Error("failed to find user by token")
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryData) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error; err != nil {
		r.log.WithError(err).Error("failed to delete user")
		return err
	}
	return nil
}

func (r *userRepositoryData) FindAll(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		r.log.WithError(err).Error("failed to find all users")
		return nil, err
	}
	return users, nil
}

func (r *userRepositoryData) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("name = ?", username).First(&user).Error
	return &user, err
}
