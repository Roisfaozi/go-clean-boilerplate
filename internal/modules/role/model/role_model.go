package model

// RoleResponse defines the structure for role API responses.
type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateRoleRequest defines the structure for creating a new role.
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description,omitempty"`
}

// UpdateRoleRequest defines the structure for updating a role.
type UpdateRoleRequest struct {
	Description string `json:"description" validate:"required"`
}