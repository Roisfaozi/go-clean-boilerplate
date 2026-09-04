package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type AccessRight struct {
	ID             string                `db:"id" gorm:"primaryKey;column:id"`
	OrganizationID *string               `db:"organization_id" gorm:"column:organization_id;index:idx_access_org_deleted;index"`
	Name           string                `db:"name" gorm:"column:name;type:varchar(191);unique;not null"`
	Description    string                `db:"description" gorm:"column:description;type:text"`
	Endpoints      []Endpoint            `db:"-" gorm:"many2many:access_right_endpoints;"`
	CreatedAt      int64                 `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt      int64                 `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index;index:idx_access_org_deleted"`
}

func (a *AccessRight) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	return nil
}

func (AccessRight) TableName() string {
	return "access_rights"
}

type Endpoint struct {
	ID        string                `db:"id" gorm:"primaryKey;column:id"`
	Path      string                `db:"path" gorm:"column:path;type:varchar(191);not null"`
	Method    string                `db:"method" gorm:"column:method;type:varchar(10);not null"`
	CreatedAt int64                 `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64                 `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index"`
}

func (e *Endpoint) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	return nil
}

func (Endpoint) TableName() string {
	return "endpoints"
}

type AccessRightEndpoint struct {
	AccessRightID string `db:"access_right_id" gorm:"primaryKey;column:access_right_id"`
	EndpointID    string `db:"endpoint_id" gorm:"primaryKey;column:endpoint_id"`
}

func (AccessRightEndpoint) TableName() string {
	return "access_right_endpoints"
}
