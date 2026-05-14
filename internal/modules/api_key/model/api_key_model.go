package model

import "time"

type CreateApiKeyRequest struct {
	Name      string     `json:"name" validate:"required,min=3,max=255"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type ApiKeyResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	OrganizationID string     `json:"organization_id"`
	UserID         string     `json:"user_id"`
	Scopes         []string   `json:"scopes"`
	ExpiresAt      *time.Time `json:"expires_at"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
}

type CreateApiKeyResponse struct {
	ApiKeyResponse
	Key string `json:"api_key"` // Only returned once upon creation
}
