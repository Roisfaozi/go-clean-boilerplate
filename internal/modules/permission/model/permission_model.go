package model

type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required,max=100"`
	Role   string `json:"role" validate:"required,max=50"`
}

type GrantPermissionRequest struct {
	Role   string `json:"role" validate:"required,max=50"`
	Path   string `json:"path" validate:"required,max=200"`
	Method string `json:"method" validate:"required,max=10"`
}

type UpdatePermissionRequest struct {
	OldPermission []string `json:"old_permission" validate:"required,min=3,max=3"`
	NewPermission []string `json:"new_permission" validate:"required,min=3,max=3"`
}
