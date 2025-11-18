package repository

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
)

type IAccessRepository interface {
	CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error
	FindAccessRightByName(ctx context.Context, name string) (*entity.AccessRight, error)
	GetAllAccessRights(ctx context.Context) ([]entity.AccessRight, error)
	CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error
	GetEndpointByPathAndMethod(ctx context.Context, path, method string) (*entity.Endpoint, error)
	LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID uint) error
	UnlinkEndpointFromAccessRight(ctx context.Context, accessRightID, endpointID uint) error
	GetEndpointsForAccessRight(ctx context.Context, accessRightID uint) ([]entity.Endpoint, error)
}
