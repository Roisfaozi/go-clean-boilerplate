package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type IAccessUseCase interface {
	CreateAccessRight(ctx context.Context, req model.CreateAccessRightRequest) (*model.AccessRightResponse, error)
	GetAllAccessRights(ctx context.Context) (*model.AccessRightListResponse, error)
	GetAccessRightsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) (*model.AccessRightListResponse, error)
	DeleteAccessRight(ctx context.Context, id string) error
	CreateEndpoint(ctx context.Context, req model.CreateEndpointRequest) (*model.EndpointResponse, error)
	GetEndpointsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*model.EndpointResponse, error)
	DeleteEndpoint(ctx context.Context, id string) error
	LinkEndpointToAccessRight(ctx context.Context, req model.LinkEndpointRequest) error
}
