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
	ID        string                `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name      string                `gorm:"column:name;type:varchar(255);not null"`
	Slug      string                `gorm:"column:slug;type:varchar(100);uniqueIndex;not null"`
	OwnerID   string                `gorm:"column:owner_id;type:varchar(36);not null;index"`
	Settings  map[string]interface{} `gorm:"serializer:json"`
	Status    string                `gorm:"column:status;type:varchar(20);default:'active';index"`
	CreatedAt int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index"`

	// Relations
	Members []OrganizationMember `gorm:"foreignKey:OrganizationID;references:ID"`
}

func (Organization) TableName() string {
	return "organizations"
}
