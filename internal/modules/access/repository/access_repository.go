package repository

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
	"gorm.io/gorm"
)

// AccessRepository is the concrete implementation of IAccessRepository
type AccessRepository struct {
	db *gorm.DB
}

// NewAccessRepository creates a new AccessRepository instance.
//
// Parameters:
// - db: The database connection.
//
// Returns:
// - IAccessRepository: The newly created AccessRepository instance.
func NewAccessRepository(db *gorm.DB) IAccessRepository {
	return &AccessRepository{db: db}
}
func (r *AccessRepository) CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error {
	return r.db.WithContext(ctx).Create(accessRight).Error
}

func (r *AccessRepository) FindAccessRightByName(ctx context.Context, name string) (*entity.AccessRight, error) {
	var accessRight entity.AccessRight
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&accessRight).Error
	return &accessRight, err
}

func (r *AccessRepository) FindAccessRightByID(ctx context.Context, id uint) (*entity.AccessRight, error) {
	var accessRight entity.AccessRight
	err := r.db.WithContext(ctx).First(&accessRight, id).Error
	return &accessRight, err
}

func (r *AccessRepository) GetAllAccessRights(ctx context.Context) ([]entity.AccessRight, error) {
	var accessRights []entity.AccessRight
	err := r.db.WithContext(ctx).Find(&accessRights).Error
	return accessRights, err
}

func (r *AccessRepository) DeleteAccessRight(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.AccessRight{}, id).Error
}

func (r *AccessRepository) CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
	return r.db.WithContext(ctx).Create(endpoint).Error
}

func (r *AccessRepository) GetEndpointByPathAndMethod(ctx context.Context, path, method string) (*entity.Endpoint, error) {
	var endpoint entity.Endpoint
	err := r.db.WithContext(ctx).Where("path = ? AND method = ?", path, method).First(&endpoint).Error
	return &endpoint, err
}

func (r *AccessRepository) FindEndpointByID(ctx context.Context, id uint) (*entity.Endpoint, error) {
	var endpoint entity.Endpoint
	err := r.db.WithContext(ctx).First(&endpoint, id).Error
	return &endpoint, err
}

func (r *AccessRepository) DeleteEndpoint(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Endpoint{}, id).Error
}

func (r *AccessRepository) LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID uint) error {
	accessRight := entity.AccessRight{ID: accessRightID}
	endpoint := entity.Endpoint{ID: endpointID}
	return r.db.WithContext(ctx).Model(&accessRight).Association("Endpoints").Append(&endpoint)
}

func (r *AccessRepository) UnlinkEndpointFromAccessRight(ctx context.Context, accessRightID, endpointID uint) error {
	accessRight := entity.AccessRight{ID: accessRightID}
	endpoint := entity.Endpoint{ID: endpointID}
	return r.db.WithContext(ctx).Model(&accessRight).Association("Endpoints").Delete(&endpoint)
}

func (r *AccessRepository) GetEndpointsForAccessRight(ctx context.Context, accessRightID uint) ([]entity.Endpoint, error) {
	var accessRight entity.AccessRight
	err := r.db.WithContext(ctx).Preload("Endpoints").First(&accessRight, accessRightID).Error
	if err != nil {
		return nil, err
	}
	return accessRight.Endpoints, nil
}