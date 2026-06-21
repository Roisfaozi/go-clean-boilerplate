package model

type CreateAuditLogRequest struct {
	OrganizationID string      `json:"organization_id"`
	UserID         string      `json:"user_id" validate:"required,xss"`
	Action         string      `json:"action" validate:"required,xss"`
	Entity         string      `json:"entity" validate:"required,xss"`
	EntityID       string      `json:"entity_id" validate:"required,xss"`
	OldValues      interface{} `json:"old_values"`
	NewValues      interface{} `json:"new_values"`
	IPAddress      string      `json:"ip_address" validate:"xss"`
	UserAgent      string      `json:"user_agent" validate:"xss"`
}

type AuditLogResponse struct {
	ID             string      `json:"id"`
	OrganizationID *string     `json:"organization_id"`
	UserID         string      `json:"user_id"`
	Action         string      `json:"action"`
	Entity         string      `json:"entity"`
	EntityID       string      `json:"entity_id"`
	OldValues      interface{} `json:"old_values"`
	NewValues      interface{} `json:"new_values"`
	IPAddress      string      `json:"ip_address"`
	UserAgent      string      `json:"user_agent"`
	CreatedAt      int64       `json:"created_at"`
}

type AuditLogExportPayload struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	FromDate       string `json:"from_date"`
	ToDate         string `json:"to_date"`
	Format         string `json:"format"` // "csv" or "excel" (future)
}
