package model

// CreateOrganizationRequest represents the request to create a new organization
type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
	Slug string `json:"slug" validate:"required,min=2,max=100,slug"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	Name   string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active suspended inactive"`
}

// OrganizationResponse represents the response for organization operations
type OrganizationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	OwnerID   string `json:"owner_id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// InviteMemberRequest represents the request to invite a user to an organization
type InviteMemberRequest struct {
	Email  string `json:"email" validate:"required,email"`
	UserID string `json:"user_id,omitempty"`
	RoleID string `json:"role_id" validate:"required"`
}

// UpdateMemberRequest represents the request to update a member's role or status
type UpdateMemberRequest struct {
	RoleID string `json:"role_id,omitempty"`
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active suspended banned"`
}

// MemberResponse represents a member in an organization
type MemberResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	RoleID         string `json:"role_id"`
	Status         string `json:"status"`
	JoinedAt       int64  `json:"joined_at"`
}

// UserOrganizationsResponse represents the list of organizations a user belongs to
type UserOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Total         int                    `json:"total"`
}
