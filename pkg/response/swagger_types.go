package response

import (
	accessModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	roleModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	webhookModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
)

// SwaggerSuccessResponseWrapper is a generic success response wrapper for Swagger documentation
type SwaggerSuccessResponseWrapper struct {
	Data   interface{}   `json:"data"`
	Paging *PageMetadata `json:"paging,omitempty"`
}

type SwaggerUserResponseWrapper struct {
	Data   userModel.UserResponse `json:"data"`
	Paging *PageMetadata          `json:"paging,omitempty"`
}

type SwaggerUserListResponseWrapper struct {
	Data   []userModel.UserResponse `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
}

type SwaggerLoginResponseWrapper struct {
	Data   authModel.LoginResponse `json:"data"`
	Paging *PageMetadata           `json:"paging,omitempty"`
}

type SwaggerTokenResponseWrapper struct {
	Data   authModel.TokenResponse `json:"data"`
	Paging *PageMetadata           `json:"paging,omitempty"`
}

type SwaggerGeneralResponseWrapper struct {
	Data   map[string]string `json:"data"`
	Paging *PageMetadata     `json:"paging,omitempty"`
}

type SwaggerErrorResponseWrapper struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type SwaggerRoleResponseWrapper struct {
	Data   roleModel.RoleResponse `json:"data"`
	Paging *PageMetadata          `json:"paging,omitempty"`
}

type SwaggerRoleListResponseWrapper struct {
	Data   []roleModel.RoleResponse `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
}

type SwaggerAccessRightResponseWrapper struct {
	Data   accessModel.AccessRightResponse `json:"data"`
	Paging *PageMetadata                   `json:"paging,omitempty"`
}

type SwaggerAccessRightListResponseWrapper struct {
	Data   accessModel.AccessRightListResponse `json:"data"`
	Paging *PageMetadata                       `json:"paging,omitempty"`
}

type SwaggerEndpointResponseWrapper struct {
	Data   accessModel.EndpointResponse `json:"data"`
	Paging *PageMetadata                `json:"paging,omitempty"`
}

type SwaggerEndpointListResponseWrapper struct {
	Data   []accessModel.EndpointResponse `json:"data"`
	Paging *PageMetadata                  `json:"paging,omitempty"`
}

type SwaggerPermissionListResponseWrapper struct {
	Data   [][]string    `json:"data"`
	Paging *PageMetadata `json:"paging,omitempty"`
}

// Organization Swagger Types

type SwaggerOrganizationResponseWrapper struct {
	Data   orgModel.OrganizationResponse `json:"data"`
	Paging *PageMetadata                 `json:"paging,omitempty"`
}

type SwaggerOrganizationListResponseWrapper struct {
	Data   orgModel.UserOrganizationsResponse `json:"data"`
	Paging *PageMetadata                      `json:"paging,omitempty"`
}

// Audit Swagger Types

type SwaggerAuditLogListResponseWrapper struct {
	Data   []map[string]interface{} `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
}

// Webhook Swagger Types

type SwaggerWebhookResponseWrapper struct {
	Data   webhookModel.WebhookResponse `json:"data"`
	Paging *PageMetadata                `json:"paging,omitempty"`
}

type SwaggerWebhookListResponseWrapper struct {
	Data   []webhookModel.WebhookResponse `json:"data"`
	Paging *PageMetadata                  `json:"paging,omitempty"`
}

type SwaggerWebhookLogListResponseWrapper struct {
	Data   []map[string]interface{} `json:"data"`
	Paging *PageMetadata            `json:"paging,omitempty"`
}
