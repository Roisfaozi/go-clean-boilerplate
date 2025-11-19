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

// UpdatePermissionRequest defines the structure for updating a permission.
type UpdatePermissionRequest struct {
	OldPermission []string `json:"old_permission" validate:"required,min=3,max=3"`
	NewPermission []string `json:"new_permission" validate:"required,min=3,max=3"`
}
