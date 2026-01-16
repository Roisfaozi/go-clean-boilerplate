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

type RoleInheritanceRequest struct {
	ChildRole  string `json:"child_role" validate:"required"`
	ParentRole string `json:"parent_role" validate:"required"`
}

type PermissionCheckItem struct {
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

type BatchPermissionCheckRequest struct {
	Items []PermissionCheckItem `json:"items" validate:"required,min=1"`
}

type BatchPermissionCheckResponse struct {
	Results map[string]bool `json:"results"`
}
