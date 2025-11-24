package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/access/model"
)

type IAccessUseCase interface {
	CreateAccessRight(ctx context.Context, req model.CreateAccessRightRequest) (*model.AccessRightResponse, error)
	GetAllAccessRights(ctx context.Context) (*model.AccessRightListResponse, error)
	CreateEndpoint(ctx context.Context, req model.CreateEndpointRequest) (*model.EndpointResponse, error)
	LinkEndpointToAccessRight(ctx context.Context, req model.LinkEndpointRequest) error
}
