package response

import (
	authModel "github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
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