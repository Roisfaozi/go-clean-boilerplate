package entity

import "gorm.io/plugin/soft_delete"

type Role struct {
	ID             string                `db:"id" gorm:"type:varchar(36);primary_key"`
	Name           string                `db:"name" gorm:"type:varchar(50);not null;uniqueIndex:idx_role_name_org"`
	OrganizationID *string               `db:"organization_id" gorm:"type:varchar(36);uniqueIndex:idx_role_name_org;index:idx_role_org_deleted"`
	Description    string                `db:"description" gorm:"type:text"`
	CreatedAt      int64                 `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt      int64                 `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index;index:idx_role_org_deleted"`
}

func (Role) TableName() string {
	return "roles"
}
