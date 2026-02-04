package entity

import "gorm.io/gorm"

type AuditLog struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)"`
	OrganizationID *string   `gorm:"index;type:varchar(36)"`
	UserID    string         `gorm:"index;type:varchar(36);not null"`
	Action    string         `gorm:"size:50;not null"`
	Entity    string         `gorm:"size:50;not null"`
	EntityID  string         `gorm:"size:100;not null"`
	OldValues string         `gorm:"type:json"`
	NewValues string         `gorm:"type:json"`
	IPAddress string         `gorm:"size:45"`
	UserAgent string         `gorm:"size:255"`
	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
