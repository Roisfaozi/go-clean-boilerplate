//go:generate mockery --name RoleUseCase --output ../test/mocks --outpkg mocks
package usecase

import (
	"context"

	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
)

type RoleUseCase interface {
	Create(ctx context.Context, request *model.CreateRoleRequest) (*model.RoleResponse, error)
	GetAll(ctx context.Context) ([]model.RoleResponse, error)
	Delete(ctx context.Context, id string) error
}
