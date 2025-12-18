package model

type CreateAuditLogRequest struct {
	UserID    string      `json:"user_id" validate:"required"`
	Action    string      `json:"action" validate:"required"`
	Entity    string      `json:"entity" validate:"required"`
	EntityID  string      `json:"entity_id" validate:"required"`
	OldValues interface{} `json:"old_values"`
	NewValues interface{} `json:"new_values"`
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
}

type AuditLogResponse struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Action    string      `json:"action"`
	Entity    string      `json:"entity"`
	EntityID  string      `json:"entity_id"`
	OldValues interface{} `json:"old_values"`
	NewValues interface{} `json:"new_values"`
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
	CreatedAt int64       `json:"created_at"`
}
