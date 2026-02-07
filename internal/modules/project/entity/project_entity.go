package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Project struct {
	ID             string                `gorm:"primaryKey;column:id"`
	OrganizationID string                `gorm:"column:organization_id;index"`
	UserID         string                `gorm:"column:user_id;index"`
	Name           string                `gorm:"column:name;type:varchar(191);not null"`
	Domain         string                `gorm:"column:domain;type:varchar(191);not null"`
	Status         string                `gorm:"column:status;type:varchar(50);default:'active'"`
	CreatedAt      int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt      int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index"`
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
