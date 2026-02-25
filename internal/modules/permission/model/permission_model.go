package model

type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required,xss"`
	Role   string `json:"role" validate:"required,xss"`
	Domain string `json:"domain" validate:"omitempty,xss"`
}

type GrantPermissionRequest struct {
	Role   string `json:"role" validate:"required,xss"`
	Path   string `json:"path" validate:"required,xss"`
	Method string `json:"method" validate:"required,xss"`
	Domain string `json:"domain" validate:"omitempty,xss"`
}

type UpdatePermissionRequest struct {
	OldPermission []string `json:"old_permission" validate:"required,min=4,max=4,dive,xss"`
	NewPermission []string `json:"new_permission" validate:"required,min=4,max=4,dive,xss"`
}

type RoleInheritanceRequest struct {
	ChildRole  string `json:"child_role" validate:"required"`
	ParentRole string `json:"parent_role" validate:"required"`
	Domain     string `json:"domain" validate:"omitempty,xss"`
}

type PermissionCheckItem struct {
	Resource string `json:"resource" validate:"required,xss"`
	Action   string `json:"action" validate:"required,xss"`
}

type BatchPermissionCheckRequest struct {
	Items []PermissionCheckItem `json:"items" validate:"required,min=1"`
}

type BatchPermissionCheckResponse struct {
	Results map[string]bool `json:"results"`
}

type ResourceCRUD struct {
	Create bool `json:"create"`
	Read   bool `json:"read"`
	Update bool `json:"update"`
	Delete bool `json:"delete"`
}

type ResourcePermission struct {
	Name            string                  `json:"name"`
	BasePath        string                  `json:"base_path"`
	RolePermissions map[string]ResourceCRUD `json:"role_permissions"`
	EndpointCount   int                     `json:"endpoint_count"`
}

type ResourceAggregationResponse struct {
	Resources []ResourcePermission `json:"resources"`
}

type RoleNode struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Description          string     `json:"description,omitempty"`
	ParentID             *string    `json:"parent_id,omitempty"`
	Children             []RoleNode `json:"children,omitempty"`
	OwnPermissions       [][]string `json:"own_permissions"`
	InheritedPermissions [][]string `json:"inherited_permissions"`
	EffectivePermissions [][]string `json:"effective_permissions"`
}
type InheritanceTreeResponse struct {
	Roles []RoleNode `json:"roles"`
}
