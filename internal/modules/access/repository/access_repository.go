package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	querybuilder2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
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

func (r *accessRepository) getDB(ctx context.Context) *gorm.DB {
	if txDB, ok := tx.DBFromContext(ctx); ok {
		return txDB
	}
	return r.db.WithContext(ctx)
}

func (r *accessRepository) CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error {
	return r.getDB(ctx).Create(endpoint).Error
}

func (r *accessRepository) GetEndpoints(ctx context.Context) ([]*entity.Endpoint, error) {
	var endpoints []*entity.Endpoint
	// Endpoints are usually global, but if we want per-tenant endpoints in future:
	if err := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Find(&endpoints).Error; err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (r *accessRepository) FindEndpointsDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.Endpoint, int64, error) {
	var endpoints []*entity.Endpoint
	var total int64
	query := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Model(&entity.Endpoint{})

	query, err := querybuilder2.GenerateDynamicQuery(query, &entity.Endpoint{}, filter)
	if err != nil {
		return nil, 0, err
	}

	if !filter.SkipCount {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		total = -1
	}

	query, err = querybuilder2.GenerateDynamicSort(query, &entity.Endpoint{}, filter)
	if err != nil {
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Limit(filter.PageSize).Offset(offset)
	}

	if err := query.Find(&endpoints).Error; err != nil {
		return nil, 0, err
	}
	return endpoints, total, nil
}

func (r *accessRepository) GetEndpointByID(ctx context.Context, id string) (*entity.Endpoint, error) {
	var endpoint entity.Endpoint
	if err := r.getDB(ctx).First(&endpoint, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &endpoint, nil
}

func (r *accessRepository) DeleteEndpoint(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&entity.Endpoint{}, "id = ?", id).Error
}

func (r *accessRepository) CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error {
	return r.getDB(ctx).Create(accessRight).Error
}

func (r *accessRepository) GetAccessRights(ctx context.Context) ([]*entity.AccessRight, error) {
	var accessRights []*entity.AccessRight
	if err := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Preload("Endpoints").Find(&accessRights).Error; err != nil {
		return nil, err
	}
	return accessRights, nil
}

func (r *accessRepository) FindAccessRightsDynamic(ctx context.Context, filter *querybuilder2.DynamicFilter) ([]*entity.AccessRight, int64, error) {
	var accessRights []*entity.AccessRight
	var total int64
	query := r.getDB(ctx).Scopes(database.OrganizationScope(ctx)).Model(&entity.AccessRight{}).Preload("Endpoints")

	query, err := querybuilder2.GenerateDynamicQuery(query, &entity.AccessRight{}, filter)
	if err != nil {
		return nil, 0, err
	}

	if !filter.SkipCount {
		if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
			return nil, 0, err
		}
	} else {
		total = -1
	}

	query, err = querybuilder2.GenerateDynamicSort(query, &entity.AccessRight{}, filter)
	if err != nil {
		return nil, 0, err
	}

	if filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Limit(filter.PageSize).Offset(offset)
	}

	if err := query.Find(&accessRights).Error; err != nil {
		return nil, 0, err
	}
	return accessRights, total, nil
}

func (r *accessRepository) GetAccessRightByID(ctx context.Context, id string) (*entity.AccessRight, error) {
	var accessRight entity.AccessRight
	if err := r.getDB(ctx).Preload("Endpoints").First(&accessRight, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &accessRight, nil
}

func (r *accessRepository) DeleteAccessRight(ctx context.Context, id string) error {
	return r.getDB(ctx).Delete(&entity.AccessRight{}, "id = ?", id).Error
}

func (r *accessRepository) LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID string) error {
	return r.getDB(ctx).Model(&entity.AccessRight{ID: accessRightID}).Association("Endpoints").Append(&entity.Endpoint{ID: endpointID})
}


func (r *accessRepository) UnlinkEndpointFromAccessRight(ctx context.Context, accessRightID, endpointID string) error {
	return r.getDB(ctx).Model(&entity.AccessRight{ID: accessRightID}).Association("Endpoints").Delete(&entity.Endpoint{ID: endpointID})
}
