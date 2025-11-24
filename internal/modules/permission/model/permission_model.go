package model

type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required"`
}

type GrantPermissionRequest struct {
	Role   string `json:"role" validate:"required"`
	Path   string `json:"path" validate:"required"`
	Method string `json:"method" validate:"required"`
}

type UpdatePermissionRequest struct {
	OldPermission []string `json:"old_permission" validate:"required,min=3,max=3"`
	NewPermission []string `json:"new_permission" validate:"required,min=3,max=3"`
}
