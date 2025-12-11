package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	querybuilder2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type accessRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewAccessRepository(db *gorm.DB, log *logrus.Logger) AccessRepository {
	return &accessRepository{
		db:  db,
		log: log,
	}
}

// Endpoint Methods
func (r *accessRepository) CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
	return r.db.WithContext(ctx).Create(endpoint).Error
}

func (r *accessRepository) GetEndpoints(ctx context.Context) ([]*entity.Endpoint, error) {
	var endpoints []*entity.Endpoint
	if err := r.db.WithContext(ctx).Find(&endpoints).Error; err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (r *accessRepository) FindEndpointsDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.Endpoint, error) {
	var endpoints []*entity.Endpoint
	query := r.db.WithContext(ctx)

	where, args, _, err := querybuilder2.GenerateDynamicQuery[entity.Endpoint](filter)
	if err != nil {
		return nil, err
	}
	if where != "" {
		query = query.Where(where, args...)
	}

	sort, err := querybuilder2.GenerateDynamicSort[entity.Endpoint](filter)
	if err != nil {
		return nil, err
	}
	if sort != "" {
		query = query.Order(sort)
	}

	if err := query.Find(&endpoints).Error; err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (r *accessRepository) GetEndpointByID(ctx context.Context, id string) (*entity.Endpoint, error) {
	var endpoint entity.Endpoint
	if err := r.db.WithContext(ctx).First(&endpoint, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *accessRepository) DeleteEndpoint(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Endpoint{}, "id = ?", id).Error
}

// AccessRight Methods
func (r *accessRepository) CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error {
	return r.db.WithContext(ctx).Create(accessRight).Error
}

func (r *accessRepository) GetAccessRights(ctx context.Context) ([]*entity.AccessRight, error) {
	var accessRights []*entity.AccessRight
	if err := r.db.WithContext(ctx).Preload("Endpoints").Find(&accessRights).Error; err != nil {
		return nil, err
	}
	return accessRights, nil
}

func (r *accessRepository) FindAccessRightsDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.AccessRight, error) {
	var accessRights []*entity.AccessRight
	query := r.db.WithContext(ctx).Preload("Endpoints")

	where, args, _, err := querybuilder2.GenerateDynamicQuery[entity.AccessRight](filter)
	if err != nil {
		return nil, err
	}
	if where != "" {
		query = query.Where(where, args...)
	}

	sort, err := querybuilder2.GenerateDynamicSort[entity.AccessRight](filter)
	if err != nil {
		return nil, err
	}
	if sort != "" {
		query = query.Order(sort)
	}

	if err := query.Find(&accessRights).Error; err != nil {
		return nil, err
	}
	return accessRights, nil
}

func (r *accessRepository) GetAccessRightByID(ctx context.Context, id string) (*entity.AccessRight, error) {
	var accessRight entity.AccessRight
	if err := r.db.WithContext(ctx).Preload("Endpoints").First(&accessRight, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &accessRight, nil
}

func (r *accessRepository) DeleteAccessRight(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.AccessRight{}, "id = ?", id).Error
}

func (r *accessRepository) LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID string) error {
	return r.db.WithContext(ctx).Model(&entity.AccessRight{ID: accessRightID}).Association("Endpoints").Append(&entity.Endpoint{ID: endpointID})
}
