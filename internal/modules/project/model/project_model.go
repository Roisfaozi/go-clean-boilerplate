package model

type CreateProjectRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=100,xss"`
	Domain string `json:"domain" validate:"required,min=1,max=100,xss"`
}

type UpdateProjectRequest struct {
	Name   string `json:"name" validate:"omitempty,min=1,max=100,xss"`
	Domain string `json:"domain" validate:"omitempty,min=1,max=100,xss"`
	Status string `json:"status" validate:"omitempty"`
}

type ProjectResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	Domain         string `json:"domain"`
	Status         string `json:"status"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}
