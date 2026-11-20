package model

import (
	"strings"

	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
)

// OrganizationResponse represents the response for organization operations
type OrganizationResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	OwnerID   string                 `json:"owner_id"`
	Settings  map[string]interface{} `json:"settings"`
	Status    string                 `json:"status"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

type UserOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Total         int                    `json:"total"`
}

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100,xss"`
	Slug string `json:"slug" validate:"omitempty,min=3,max=100,slug"`
}

func (r *CreateOrganizationRequest) Sanitize() {
	r.Name = pkg.SanitizeString(r.Name)
	r.Slug = strings.ToLower(strings.TrimSpace(r.Slug))
}

type UpdateOrganizationRequest struct {
	Name     string                 `json:"name" validate:"omitempty,min=3,max=100,xss"`
	Settings map[string]interface{} `json:"settings"`
	Status   string                 `json:"status" validate:"omitempty,oneof=active suspended"`
}

func (r *UpdateOrganizationRequest) Sanitize() {
	if r.Name != "" {
		r.Name = pkg.SanitizeString(r.Name)
	}
	if r.Settings != nil {
		for k, v := range r.Settings {
			if strVal, ok := v.(string); ok {
				r.Settings[k] = pkg.SanitizeString(strVal)
			}
		}
	}
}

type InviteMemberRequest struct {
	Email  string `json:"email" validate:"required,email"`
	UserID string `json:"user_id"` // Optional, if inviting generic email
	RoleID string `json:"role_id" validate:"required"`
}

type AcceptInvitationRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password"` // Required for new users
	Name     string `json:"name"`     // Optional
}

type UpdateMemberRequest struct {
	RoleID string `json:"role_id"`
	Status string `json:"status" validate:"omitempty,oneof=active suspended"`
}

type MemberResponse struct {
	ID             string                  `json:"id"`
	OrganizationID string                  `json:"organization_id"`
	UserID         string                  `json:"user_id"`
	User           *userModel.UserResponse `json:"user,omitempty"`
	RoleID         string                  `json:"role_id"`
	Status         string                  `json:"status"`
	JoinedAt       int64                   `json:"joined_at"`
}
