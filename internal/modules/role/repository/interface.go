package repository

import (
	"context"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/entity"
)

//go:generate mockery --name RoleRepository --output ../test/mocks --outpkg mocks
// RoleRepository defines the interface for role repository operations.
type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) error
	FindByID(ctx context.Context, id string) (*entity.Role, error)
	FindByName(ctx context.Context, name string) (*entity.Role, error)
	FindAll(ctx context.Context) ([]*entity.Role, error)
	Delete(ctx context.Context, id string) error
}