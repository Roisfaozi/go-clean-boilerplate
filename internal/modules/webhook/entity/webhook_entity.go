package entity

type Webhook struct {
	ID             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	OrganizationID string `db:"organization_id" json:"organization_id"`
	URL            string `db:"url" json:"url"`
	Events         string `db:"events" json:"events"` // Stored as JSON string
	Secret         string `db:"secret" json:"secret"`
	IsActive       bool   `db:"is_active" json:"is_active"`
	CreatedAt      int64  `db:"created_at" json:"created_at"`
	UpdatedAt      int64  `db:"updated_at" json:"updated_at"`
	DeletedAt      int64  `db:"deleted_at" json:"-"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

type WebhookLog struct {
	ID                 string `db:"id" json:"id"`
	WebhookID          string `db:"webhook_id" json:"webhook_id"`
	EventType          string `db:"event_type" json:"event_type"`
	Payload            string `db:"payload" json:"payload"`
	ResponseStatusCode int    `db:"response_status_code" json:"response_status_code"`
	ResponseBody       string `db:"response_body" json:"response_body"`
	ExecutionTime      int64  `db:"execution_time" json:"execution_time"`
	ErrorMessage       string `db:"error_message" json:"error_message"`
	RetryCount         int    `db:"retry_count" json:"retry_count"`
	CreatedAt          int64  `db:"created_at" json:"created_at"`
}

func (WebhookLog) TableName() string {
	return "webhook_logs"
}
