package entity

import (
	"gorm.io/plugin/soft_delete"
)

type AuditLog struct {
	ID             string                `gorm:"primaryKey;type:varchar(36)"`
	OrganizationID *string               `gorm:"index:idx_audit_org_deleted;index;type:varchar(36)"`
	UserID         string                `gorm:"index:idx_audit_user_deleted;index;type:varchar(36);not null"`
	Action         string                `gorm:"size:50;not null"`
	Entity         string                `gorm:"size:50;not null"`
	EntityID       string                `gorm:"size:100;not null"`
	OldValues      string                `gorm:"type:json"`
	NewValues      string                `gorm:"type:json"`
	IPAddress      string                `gorm:"size:45"`
	UserAgent      string                `gorm:"size:255"`
	CreatedAt      int64                 `gorm:"autoCreateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;index:idx_audit_org_deleted;index:idx_audit_user_deleted"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
