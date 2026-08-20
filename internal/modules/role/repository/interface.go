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
	FindByNameInScope(ctx context.Context, name string, orgID *string) (*entity.Role, error)
	FindByID(ctx context.Context, id string) (*entity.Role, error)
	FindOrganizationRoleByID(ctx context.Context, organizationID, roleID string) (*entity.Role, error)
	FindOrganizationRoles(ctx context.Context, organizationID string) ([]*entity.Role, error)
	Create(ctx context.Context, role *entity.Role) error
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id string) error
	DeleteInOrg(ctx context.Context, orgID, roleID string) error
}
