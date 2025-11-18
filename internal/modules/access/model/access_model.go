package model

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/access/entity"
)

// CreateAccessRightRequest represents the request payload for creating a new access right
type CreateAccessRightRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=255"`
}

// CreateEndpointRequest defines the payload for creating a new endpoint.
type CreateEndpointRequest struct {
	Path   string `json:"path" validate:"required,min=1,max=191"`
	Method string `json:"method" validate:"required,min=1,max=10"`
}

// LinkEndpointRequest defines the payload for linking an endpoint to an access right.
type LinkEndpointRequest struct {
	AccessRightID uint `json:"access_right_id" validate:"required"`
	EndpointID    uint `json:"endpoint_id" validate:"required"`
}

// UpdateAccessRightRequest represents the request payload for updating an access right
type UpdateAccessRightRequest struct {
	Name        string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description string `json:"description,omitempty" validate:"max=255"`
}

// AddEndpointToAccessRightRequest represents the request payload for adding an endpoint to an access right
type AddEndpointToAccessRightRequest struct {
	EndpointID uint `json:"endpoint_id" validate:"required"`
}

// AccessRightResponse represents the access right response structure
type AccessRightResponse struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Endpoints   []EndpointResponse `json:"endpoints,omitempty"`
	CreatedAt   int64              `json:"created_at"`
	UpdatedAt   int64              `json:"updated_at"`
}

// EndpointResponse represents the endpoint response structure
type EndpointResponse struct {
	ID        uint   `json:"id"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	CreatedAt int64  `json:"created_at"`
}

// AccessRightListResponse represents a list of access rights
type AccessRightListResponse struct {
	Data []AccessRightResponse `json:"data"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ConvertAccessRightToResponse converts an entity.AccessRight to AccessRightResponse
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

// ConvertAccessRightListToResponse converts a slice of entity.AccessRight to AccessRightListResponse
func ConvertAccessRightListToResponse(accessRights []entity.AccessRight) *AccessRightListResponse {
	response := &AccessRightListResponse{
		Data: make([]AccessRightResponse, 0, len(accessRights)),
	}

	for _, ar := range accessRights {
		response.Data = append(response.Data, *ConvertAccessRightToResponse(&ar))
	}

	response.Meta.Total = len(accessRights)
	return response
}
