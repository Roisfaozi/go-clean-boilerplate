package model

import (
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
)

type RoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,max=50,xss"`
	Description string `json:"description,omitempty" validate:"omitempty,xss"`
}

type UpdateRoleRequest struct {
	Description string `json:"description" validate:"required,xss"`
}

func (r *CreateRoleRequest) Sanitize() {
	r.Name = validation.SanitizeString(strings.TrimSpace(r.Name))
	r.Description = validation.SanitizeString(strings.TrimSpace(r.Description))
}

func (r *UpdateRoleRequest) Sanitize() {
	r.Description = validation.SanitizeString(strings.TrimSpace(r.Description))
}
