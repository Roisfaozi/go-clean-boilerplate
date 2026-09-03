package entity

import (
	"gorm.io/plugin/soft_delete"
)

type ApiKey struct {
	ID             string                `gorm:"type:varchar(36);primaryKey"`
	Name           string                `gorm:"type:varchar(255);not null"`
	KeyHash        string                `gorm:"type:varchar(255);not null;uniqueIndex"`
	OrganizationID string                `gorm:"type:varchar(36);not null;index"`
	UserID         string                `gorm:"type:varchar(36);not null;index"`
	Scopes         string                `gorm:"type:text"`
	ExpiresAt      *int64                `gorm:"type:bigint"`
	LastUsedAt     *int64                `gorm:"type:bigint"`
	IsActive       bool                  `gorm:"type:boolean;default:true"`
	CreatedAt      int64                 `gorm:"type:bigint;autoCreateTime:milli"`
	UpdatedAt      int64                 `gorm:"type:bigint;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `gorm:"type:bigint;not null;default:0;softDelete:milli"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
