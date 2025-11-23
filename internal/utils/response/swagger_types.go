package response

import (
	accessModel "github.com/Roisfaozi/casbin-db/internal/modules/access/model"
	authModel "github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	roleModel "github.com/Roisfaozi/casbin-db/internal/modules/role/model"
	userModel "github.com/Roisfaozi/casbin-db/internal/modules/user/model"
)

// SwaggerUserResponseWrapper wraps the UserResponse for Swagger documentation.
type SwaggerUserResponseWrapper struct {
	Data   userModel.UserResponse `json:"data"`
	Paging *PageMetadata          `json:"paging,omitempty"`
	Errors string                 `json:"errors,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// SwaggerUserListResponseWrapper wraps a list of UserResponse for Swagger documentation.
type SwaggerUserListResponseWrapper struct {
	Data   []userModel.UserResponse `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
	Errors string                   `json:"errors,omitempty"`
	Error  string                   `json:"error,omitempty"`
}

// SwaggerLoginResponseWrapper wraps the LoginResponse for Swagger documentation.
type SwaggerLoginResponseWrapper struct {
	Data   authModel.LoginResponse `json:"data"`
	Paging *PageMetadata           `json:"paging,omitempty"`
	Errors string                  `json:"errors,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

// SwaggerTokenResponseWrapper wraps the TokenResponse for Swagger documentation.
type SwaggerTokenResponseWrapper struct {
	Data   authModel.TokenResponse `json:"data"`
	Paging *PageMetadata           `json:"paging,omitempty"`
	Errors string                  `json:"errors,omitempty"`
	Error  string                  `json:"error,omitempty"`
}

// SwaggerGeneralResponseWrapper wraps a generic message response (e.g. for Delete or Logout).
type SwaggerGeneralResponseWrapper struct {
	Data   map[string]string `json:"data"` // Example: {"message": "success"}
	Paging *PageMetadata     `json:"paging,omitempty"`
	Errors string            `json:"errors,omitempty"`
	Error  string            `json:"error,omitempty"`
}

// SwaggerErrorResponseWrapper wraps an error response for Swagger documentation.
type SwaggerErrorResponseWrapper struct {
	Data   interface{}   `json:"data,omitempty"`
	Paging *PageMetadata `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
	Error  string        `json:"error,omitempty"`
}

// SwaggerRoleResponseWrapper wraps the RoleResponse for Swagger documentation.
type SwaggerRoleResponseWrapper struct {
	Data   roleModel.RoleResponse `json:"data"`
	Paging *PageMetadata          `json:"paging,omitempty"`
	Errors string                 `json:"errors,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// SwaggerRoleListResponseWrapper wraps a list of RoleResponse for Swagger documentation.
type SwaggerRoleListResponseWrapper struct {
	Data   []roleModel.RoleResponse `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
	Errors string                   `json:"errors,omitempty"`
	Error  string                   `json:"error,omitempty"`
}

// SwaggerAccessRightResponseWrapper wraps the AccessRightResponse for Swagger documentation.
type SwaggerAccessRightResponseWrapper struct {
	Data   accessModel.AccessRightResponse `json:"data"`
	Paging *PageMetadata                   `json:"paging,omitempty"`
	Errors string                          `json:"errors,omitempty"`
	Error  string                          `json:"error,omitempty"`
}

// SwaggerAccessRightListResponseWrapper wraps the AccessRightListResponse for Swagger documentation.
type SwaggerAccessRightListResponseWrapper struct {
	Data   accessModel.AccessRightListResponse `json:"data"`
	Paging *PageMetadata                       `json:"paging,omitempty"`
	Errors string                              `json:"errors,omitempty"`
	Error  string                              `json:"error,omitempty"`
}

// SwaggerEndpointResponseWrapper wraps the EndpointResponse for Swagger documentation.
type SwaggerEndpointResponseWrapper struct {
	Data   accessModel.EndpointResponse `json:"data"`
	Paging *PageMetadata                `json:"paging,omitempty"`
	Errors string                       `json:"errors,omitempty"`
	Error  string                       `json:"error,omitempty"`
}

// SwaggerPermissionListResponseWrapper wraps a list of permissions (string arrays) for Swagger documentation.
type SwaggerPermissionListResponseWrapper struct {
	Data   [][]string    `json:"data"`
	Paging *PageMetadata `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
	Error  string        `json:"error,omitempty"`
}