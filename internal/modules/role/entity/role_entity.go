package entity

import "gorm.io/plugin/soft_delete"

type Role struct {
	ID             string                `gorm:"type:varchar(36);primary_key"`
	Name           string                `gorm:"type:varchar(50);not null;uniqueIndex:idx_role_name_org"`
	OrganizationID *string               `gorm:"type:varchar(36);uniqueIndex:idx_role_name_org;index:idx_role_org_deleted"`
	Description    string                `gorm:"type:text"`
	CreatedAt      int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt      int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;index:idx_role_org_deleted"`
}
