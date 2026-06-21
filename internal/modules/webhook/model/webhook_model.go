package model

type CreateWebhookRequest struct {
	Name           string   `json:"name" validate:"required,min=3,max=255"`
	OrganizationID string   `json:"organization_id" validate:"required"`
	URL            string   `json:"url" validate:"required,url"`
	Events         []string `json:"events" validate:"required,min=1"`
	Secret         string   `json:"secret" validate:"required,min=8"`
}

type UpdateWebhookRequest struct {
	Name     *string   `json:"name" validate:"omitempty,min=3,max=255"`
	URL      *string   `json:"url" validate:"omitempty,url"`
	Events   *[]string `json:"events" validate:"omitempty,min=1"`
	Secret   *string   `json:"secret" validate:"omitempty,min=8"`
	IsActive *bool     `json:"is_active"`
}

type WebhookResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	OrganizationID string   `json:"organization_id"`
	URL            string   `json:"url"`
	Events         []string `json:"events"`
	IsActive       bool     `json:"is_active"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
}

type TriggerWebhookRequest struct {
	OrganizationID string      `json:"organization_id"`
	EventType      string      `json:"event_type"`
	Payload        interface{} `json:"payload"`
}
