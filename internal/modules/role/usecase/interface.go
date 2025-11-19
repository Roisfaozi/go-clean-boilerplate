package usecase

import (
	"context"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
)

// RoleUseCase defines the interface for role use case operations.
type RoleUseCase interface {
	Create(ctx context.Context, request *model.CreateRoleRequest) (*model.RoleResponse, error)
	GetAll(ctx context.Context) ([]model.RoleResponse, error)
}
