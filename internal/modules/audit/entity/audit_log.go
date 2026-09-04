package entity

import (
	"gorm.io/plugin/soft_delete"
)

type AuditLog struct {
	ID             string                `db:"id" gorm:"primaryKey;type:varchar(36)"`
	OrganizationID *string               `db:"organization_id" gorm:"index:idx_audit_org_deleted;index;type:varchar(36)"`
	UserID         string                `db:"user_id" gorm:"index:idx_audit_user_deleted;index;type:varchar(36);not null"`
	Action         string                `db:"action" gorm:"size:50;not null"`
	Entity         string                `db:"entity" gorm:"size:50;not null"`
	EntityID       string                `db:"entity_id" gorm:"size:100;not null"`
	OldValues      string                `db:"old_values" gorm:"type:json"`
	NewValues      string                `db:"new_values" gorm:"type:json"`
	IPAddress      string                `db:"ip_address" gorm:"size:45"`
	UserAgent      string                `db:"user_agent" gorm:"size:255"`
	CreatedAt      int64                 `db:"created_at" gorm:"autoCreateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index;index:idx_audit_org_deleted;index:idx_audit_user_deleted"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
