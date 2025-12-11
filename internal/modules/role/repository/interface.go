package repository

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
)

type RoleRepository interface {
	FindAll(ctx context.Context) ([]*entity.Role, error)
	FindAllDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]*entity.Role, error)
	FindByName(ctx context.Context, name string) (*entity.Role, error)
	FindByID(ctx context.Context, id string) (*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error
}
