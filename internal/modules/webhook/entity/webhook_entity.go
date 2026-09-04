package entity

import (
	"gorm.io/plugin/soft_delete"
)

type Webhook struct {
	ID             string                `db:"id" gorm:"primaryKey;column:id;type:varchar(36)" json:"id"`
	Name           string                `db:"name" gorm:"column:name;type:varchar(255);not null" json:"name"`
	OrganizationID string                `db:"organization_id" gorm:"column:organization_id;type:varchar(36);not null;index" json:"organization_id"`
	URL            string                `db:"url" gorm:"column:url;type:text;not null" json:"url"`
	Events         string                `db:"events" gorm:"column:events;type:text;not null" json:"events"` // Stored as JSON string
	Secret         string                `db:"secret" gorm:"column:secret;type:varchar(255);not null" json:"secret"`
	IsActive       bool                  `db:"is_active" gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt      int64                 `db:"created_at" gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt      int64                 `db:"updated_at" gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
	DeletedAt      soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index" json:"-"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

type WebhookLog struct {
	ID                 string `db:"id" gorm:"primaryKey;column:id;type:varchar(36)" json:"id"`
	WebhookID          string `db:"webhook_id" gorm:"column:webhook_id;type:varchar(36);not null;index" json:"webhook_id"`
	EventType          string `db:"event_type" gorm:"column:event_type;type:varchar(255);not null" json:"event_type"`
	Payload            string `db:"payload" gorm:"column:payload;type:longtext;not null" json:"payload"`
	ResponseStatusCode int    `db:"response_status_code" gorm:"column:response_status_code;type:int" json:"response_status_code"`
	ResponseBody       string `db:"response_body" gorm:"column:response_body;type:longtext" json:"response_body"`
	ExecutionTime      int64  `db:"execution_time" gorm:"column:execution_time;type:bigint" json:"execution_time"`
	ErrorMessage       string `db:"error_message" gorm:"column:error_message;type:text" json:"error_message"`
	RetryCount         int    `db:"retry_count" gorm:"column:retry_count;default:0" json:"retry_count"`
	CreatedAt          int64  `db:"created_at" gorm:"column:created_at;autoCreateTime:milli;index" json:"created_at"`
}

func (WebhookLog) TableName() string {
	return "webhook_logs"
}
