package entity

import (
	"gorm.io/gorm"
)

type Webhook struct {
	ID             string         `gorm:"primaryKey;column:id;type:varchar(36)" json:"id"`
	Name           string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	OrganizationID string         `gorm:"column:organization_id;type:varchar(36);not null;index" json:"organization_id"`
	URL            string         `gorm:"column:url;type:text;not null" json:"url"`
	Events         string         `gorm:"column:events;type:text;not null" json:"events"` // Stored as JSON string
	Secret         string         `gorm:"column:secret;type:varchar(255);not null" json:"secret"`
	IsActive       bool           `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt      int64          `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt      int64          `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

type WebhookLog struct {
	ID                 string `gorm:"primaryKey;column:id;type:varchar(36)" json:"id"`
	WebhookID          string `gorm:"column:webhook_id;type:varchar(36);not null;index" json:"webhook_id"`
	EventType          string `gorm:"column:event_type;type:varchar(255);not null" json:"event_type"`
	Payload            string `gorm:"column:payload;type:longtext;not null" json:"payload"`
	ResponseStatusCode int    `gorm:"column:response_status_code;type:int" json:"response_status_code"`
	ResponseBody       string `gorm:"column:response_body;type:longtext" json:"response_body"`
	ExecutionTime      int64  `gorm:"column:execution_time;type:bigint" json:"execution_time"`
	ErrorMessage       string `gorm:"column:error_message;type:text" json:"error_message"`
	RetryCount         int    `gorm:"column:retry_count;default:0" json:"retry_count"`
	CreatedAt          int64  `gorm:"column:created_at;autoCreateTime:milli;index" json:"created_at"`
}
