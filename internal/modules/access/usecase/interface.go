//go:generate mockery --name IAccessUseCase --output ../test/mocks --outpkg mocks
package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
)

type IAccessUseCase interface {
	CreateAccessRight(ctx context.Context, req model.CreateAccessRightRequest) (*model.AccessRightResponse, error)
	GetAllAccessRights(ctx context.Context) (*model.AccessRightListResponse, error)
	DeleteAccessRight(ctx context.Context, id uint) error // New
	CreateEndpoint(ctx context.Context, req model.CreateEndpointRequest) (*model.EndpointResponse, error)
	DeleteEndpoint(ctx context.Context, id uint) error // New
	LinkEndpointToAccessRight(ctx context.Context, req model.LinkEndpointRequest) error
}
