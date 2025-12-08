//go:generate mockery --name IAccessRepository --output ../test/mocks --outpkg mocks
package repository

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
)

type IAccessRepository interface {
	CreateAccessRight(ctx context.Context, accessRight *entity.AccessRight) error
	FindAccessRightByName(ctx context.Context, name string) (*entity.AccessRight, error)
	FindAccessRightByID(ctx context.Context, id uint) (*entity.AccessRight, error) // NEW
	GetAllAccessRights(ctx context.Context) ([]entity.AccessRight, error)
	DeleteAccessRight(ctx context.Context, id uint) error
	CreateEndpoint(ctx context.Context, endpoint *entity.Endpoint) error
	GetEndpointByPathAndMethod(ctx context.Context, path, method string) (*entity.Endpoint, error)
	FindEndpointByID(ctx context.Context, id uint) (*entity.Endpoint, error) // NEW
	DeleteEndpoint(ctx context.Context, id uint) error
	LinkEndpointToAccessRight(ctx context.Context, accessRightID, endpointID uint) error
	UnlinkEndpointFromAccessRight(ctx context.Context, accessRightID, endpointID uint) error
	GetEndpointsForAccessRight(ctx context.Context, accessRightID uint) ([]entity.Endpoint, error)
}