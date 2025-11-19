package converter

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/role/entity"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
)

// RoleToResponse converts a role entity to a role API response model.
func RoleToResponse(role *entity.Role) *model.RoleResponse {
	return &model.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
}

// RolesToResponse converts a slice of role entities to a slice of role API response models.
func RolesToResponse(roles []*entity.Role) []model.RoleResponse {
	var roleResponses []model.RoleResponse
	for _, r := range roles {
		roleResponses = append(roleResponses, *RoleToResponse(r))
	}
	return roleResponses
}