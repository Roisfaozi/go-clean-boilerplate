package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type AccessRepository interface {
	CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error
	GetEndpoints(ctx context.Context) ([]*entity.Endpoint, error)
	FindEndpointsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.Endpoint, error)
	GetEndpointByID(ctx context.Context, id string) (*entity.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error

	CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error
	GetAccessRights(ctx context.Context) ([]*entity.AccessRight, error)
	FindAccessRightsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.AccessRight, error)
	GetAccessRightByID(ctx context.Context, id string) (*entity.AccessRight, error)
	DeleteAccessRight(ctx context.Context, id string) error

	LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID string) error
}
