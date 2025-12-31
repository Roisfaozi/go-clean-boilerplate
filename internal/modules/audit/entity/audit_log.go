package entity

import "gorm.io/gorm"

// AuditLog represents an audit trail record for business actions.
type AuditLog struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `gorm:"index;type:varchar(36);not null"`
	Action    string         `gorm:"size:50;not null"`  // e.g. CREATE, UPDATE, DELETE
	Entity    string         `gorm:"size:50;not null"`  // e.g. User, Role
	EntityID  string         `gorm:"size:100;not null"` // e.g. user-uuid
	OldValues string         `gorm:"type:json"`         // Snapshot of data before change
	NewValues string         `gorm:"type:json"`         // Snapshot of data after change
	IPAddress string         `gorm:"size:45"`
	UserAgent string         `gorm:"size:255"`
	CreatedAt int64          `gorm:"autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
