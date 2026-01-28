package model

import (
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
)

type CreateAccessRightRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100,xss"`
	Description string `json:"description" validate:"max=255,xss"`
}

type CreateEndpointRequest struct {
	Path   string `json:"path" validate:"required,min=1,max=191,xss"`
	Method string `json:"method" validate:"required,min=1,max=10,xss"`
}

type LinkEndpointRequest struct {
	AccessRightID string `json:"access_right_id" validate:"required"`
	EndpointID    string `json:"endpoint_id" validate:"required"`
}

type UpdateAccessRightRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=3,max=100,xss"`
	Description string `json:"description,omitempty" validate:"max=255,xss"`
}

type AddEndpointToAccessRightRequest struct {
	EndpointID string `json:"endpoint_id" validate:"required"`
}

type AccessRightResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Endpoints   []EndpointResponse `json:"endpoints,omitempty"`
	CreatedAt   int64              `json:"created_at"`
	UpdatedAt   int64              `json:"updated_at"`
}

type EndpointResponse struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	CreatedAt int64  `json:"created_at"`
}

type AccessRightListResponse struct {
	Data []AccessRightResponse `json:"data"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (r *CreateAccessRightRequest) Sanitize() {
	r.Name = validation.SanitizeString(strings.TrimSpace(r.Name))
	r.Description = validation.SanitizeString(strings.TrimSpace(r.Description))
}

func (r *CreateEndpointRequest) Sanitize() {
	r.Path = validation.SanitizeString(strings.TrimSpace(r.Path))
	r.Method = validation.SanitizeString(strings.TrimSpace(r.Method))
}

func (r *UpdateAccessRightRequest) Sanitize() {
	r.Name = validation.SanitizeString(strings.TrimSpace(r.Name))
	r.Description = validation.SanitizeString(strings.TrimSpace(r.Description))
}

func ConvertAccessRightToResponse(accessRight *entity.AccessRight) *AccessRightResponse {
	if accessRight == nil {
		return nil
	}

	var endpoints []EndpointResponse
	for _, ep := range accessRight.Endpoints {
		endpoints = append(endpoints, EndpointResponse{
			ID:        ep.ID,
			Path:      ep.Path,
			Method:    ep.Method,
			CreatedAt: ep.CreatedAt,
		})
	}

	return &AccessRightResponse{
		ID:          accessRight.ID,
		Name:        accessRight.Name,
		Description: accessRight.Description,
		Endpoints:   endpoints,
		CreatedAt:   accessRight.CreatedAt,
		UpdatedAt:   accessRight.UpdatedAt,
	}
}

func ConvertAccessRightListToResponse(accessRights []*entity.AccessRight) *AccessRightListResponse {
	response := &AccessRightListResponse{
		Data: make([]AccessRightResponse, 0, len(accessRights)),
	}

	for _, ar := range accessRights {
		response.Data = append(response.Data, *ConvertAccessRightToResponse(ar))
	}

	response.Meta.Total = len(accessRights)
	return response
}
