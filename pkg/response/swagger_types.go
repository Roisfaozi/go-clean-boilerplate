package response

import (
	accessModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	roleModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
)

// SwaggerUserResponseWrapper wraps the UserResponse for Swagger documentation (Success).
type SwaggerUserResponseWrapper struct {
	Data   userModel.UserResponse `json:"data"`
	Paging *PageMetadata          `json:"paging,omitempty"`
}

// SwaggerUserListResponseWrapper wraps a list of UserResponse for Swagger documentation (Success).
type SwaggerUserListResponseWrapper struct {
	Data   []userModel.UserResponse `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
}

// SwaggerLoginResponseWrapper wraps the LoginResponse for Swagger documentation (Success).
type SwaggerLoginResponseWrapper struct {
	Data   authModel.LoginResponse `json:"data"`
	Paging *PageMetadata           `json:"paging,omitempty"`
}

// SwaggerTokenResponseWrapper wraps the TokenResponse for Swagger documentation (Success).
type SwaggerTokenResponseWrapper struct {
	Data   authModel.TokenResponse `json:"data"`
	Paging *PageMetadata           `json:"paging,omitempty"`
}

// SwaggerGeneralResponseWrapper wraps a generic message response (Success).
type SwaggerGeneralResponseWrapper struct {
	Data   map[string]string `json:"data"` // Example: {"message": "success"}
	Paging *PageMetadata     `json:"paging,omitempty"`
}

// SwaggerErrorResponseWrapper wraps an error response for Swagger documentation (Failure).
type SwaggerErrorResponseWrapper struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// SwaggerRoleResponseWrapper wraps the RoleResponse for Swagger documentation (Success).
type SwaggerRoleResponseWrapper struct {
	Data   roleModel.RoleResponse `json:"data"`
	Paging *PageMetadata          `json:"paging,omitempty"`
}

// SwaggerRoleListResponseWrapper wraps a list of RoleResponse for Swagger documentation (Success).
type SwaggerRoleListResponseWrapper struct {
	Data   []roleModel.RoleResponse `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
}

// SwaggerAccessRightResponseWrapper wraps the AccessRightResponse for Swagger documentation (Success).
type SwaggerAccessRightResponseWrapper struct {
	Data   accessModel.AccessRightResponse `json:"data"`
	Paging *PageMetadata                   `json:"paging,omitempty"`
}

// SwaggerAccessRightListResponseWrapper wraps the AccessRightListResponse for Swagger documentation (Success).
type SwaggerAccessRightListResponseWrapper struct {
	Data   accessModel.AccessRightListResponse `json:"data"`
	Paging *PageMetadata                       `json:"paging,omitempty"`
}

// SwaggerEndpointResponseWrapper wraps the EndpointResponse for Swagger documentation (Success).
type SwaggerEndpointResponseWrapper struct {
	Data   accessModel.EndpointResponse `json:"data"`
	Paging *PageMetadata                `json:"paging,omitempty"`
}

// SwaggerEndpointListResponseWrapper wraps a list of EndpointResponse for Swagger documentation (Success).
type SwaggerEndpointListResponseWrapper struct {
	Data   []accessModel.EndpointResponse `json:"data"`
	Paging *PageMetadata                  `json:"paging,omitempty"`
}

// SwaggerPermissionListResponseWrapper wraps a list of permissions for Swagger documentation (Success).
type SwaggerPermissionListResponseWrapper struct {
	Data   [][]string    `json:"data"`
	Paging *PageMetadata `json:"paging,omitempty"`
}
