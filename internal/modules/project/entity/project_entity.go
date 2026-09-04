package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Project struct {
	ID             string                `db:"id" gorm:"primaryKey;column:id"`
	OrganizationID string                `db:"organization_id" gorm:"column:organization_id;index:idx_project_org_deleted;index"`
	UserID         string                `db:"user_id" gorm:"column:user_id;index:idx_project_user_deleted;index"`
	Name           string                `db:"name" gorm:"column:name;type:varchar(191);not null"`
	Domain         string                `db:"domain" gorm:"column:domain;type:varchar(191);not null"`
	Status         string                `db:"status" gorm:"column:status;type:varchar(50);default:'active'"`
	CreatedAt      int64                 `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt      int64                 `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index;index:idx_project_org_deleted;index:idx_project_user_deleted"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}

func (Project) TableName() string {
	return "projects"
}
