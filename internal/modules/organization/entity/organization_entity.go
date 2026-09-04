package entity

import "gorm.io/plugin/soft_delete"

const (
	OrgStatusActive    = "active"
	OrgStatusSuspended = "suspended"
	OrgStatusInactive  = "inactive"
)

// Organization represents a tenant/workspace in the multi-tenant system.
// Users can belong to multiple organizations with different roles.
type Organization struct {
	ID        string                 `db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	Name      string                 `db:"name" gorm:"column:name;type:varchar(255);not null"`
	Slug      string                 `db:"slug" gorm:"column:slug;type:varchar(100);uniqueIndex;not null"`
	OwnerID   string                 `db:"owner_id" gorm:"column:owner_id;type:varchar(36);not null;index:idx_org_owner_deleted;index"`
	Settings  map[string]interface{} `db:"-" gorm:"serializer:json"`
	Status    string                 `db:"status" gorm:"column:status;type:varchar(20);default:'active';index:idx_org_status_deleted;index"`
	CreatedAt int64                  `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64                  `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt  `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index;index:idx_org_owner_deleted;index:idx_org_status_deleted"`

	// Relations
	Members []OrganizationMember `db:"-" gorm:"foreignKey:OrganizationID;references:ID"`
}

func (Organization) TableName() string {
	return "organizations"
}
