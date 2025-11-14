package model

// AssignRoleRequest defines the structure for assigning a role to a user.
type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required"`
}

// GrantPermissionRequest defines the structure for granting a permission to a role.
type GrantPermissionRequest struct {
	Role   string `json:"role" validate:"required"`
	Path   string `json:"path" validate:"required"`
	Method string `json:"method" validate:"required"`
}
