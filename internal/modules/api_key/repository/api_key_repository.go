package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	"gorm.io/gorm"
)

type ApiKeyRepository interface {
	Create(ctx context.Context, apiKey *entity.ApiKey) error
	FindByHash(ctx context.Context, keyHash string) (*entity.ApiKey, error)
	FindByID(ctx context.Context, id string) (*entity.ApiKey, error)
	ListByOrg(ctx context.Context, orgID string) ([]*entity.ApiKey, error)
	Update(ctx context.Context, apiKey *entity.ApiKey) error
	Delete(ctx context.Context, id string) error
}

type apiKeyRepository struct {
	db *gorm.DB
}

func NewApiKeyRepository(db *gorm.DB) ApiKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *entity.ApiKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

func (r *apiKeyRepository) FindByHash(ctx context.Context, keyHash string) (*entity.ApiKey, error) {
	var apiKey entity.ApiKey
	err := r.db.WithContext(ctx).Where("key_hash = ? AND is_active = ?", keyHash, true).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) FindByID(ctx context.Context, id string) (*entity.ApiKey, error) {
	var apiKey entity.ApiKey
	err := r.db.WithContext(ctx).First(&apiKey, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *apiKeyRepository) ListByOrg(ctx context.Context, orgID string) ([]*entity.ApiKey, error) {
	var apiKeys []*entity.ApiKey
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&apiKeys).Error
	return apiKeys, err
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *entity.ApiKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

func (r *apiKeyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ApiKey{}, "id = ?", id).Error
}
