package model

type CreateApiKeyRequest struct {
	Name      string   `json:"name" validate:"required,min=3,max=255"`
	Scopes    []string `json:"scopes"`
	ExpiresAt *int64   `json:"expires_at"`
}

type ApiKeyResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	OrganizationID string   `json:"organization_id"`
	UserID         string   `json:"user_id"`
	Scopes         []string `json:"scopes"`
	ExpiresAt      *int64   `json:"expires_at"`
	LastUsedAt     *int64   `json:"last_used_at"`
	IsActive       bool     `json:"is_active"`
	CreatedAt      int64    `json:"created_at"`
}

type CreateApiKeyResponse struct {
	ApiKeyResponse
	Key string `json:"api_key"` // Only returned once upon creation
}

type ApiKeyIdentity struct {
	ApiKeyID       string   `json:"api_key_id"`
	UserID         string   `json:"user_id"`
	OrganizationID string   `json:"organization_id"`
	Username       string   `json:"username"`
	Scopes         []string `json:"scopes"`
	ExpiresAt      *int64   `json:"expires_at"`
}
