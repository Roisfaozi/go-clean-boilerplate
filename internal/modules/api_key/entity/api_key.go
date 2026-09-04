package entity

import (
	"gorm.io/plugin/soft_delete"
)

type ApiKey struct {
	ID             string                `db:"id" gorm:"type:varchar(36);primaryKey"`
	Name           string                `db:"name" gorm:"type:varchar(255);not null"`
	KeyHash        string                `db:"key_hash" gorm:"type:varchar(255);not null;uniqueIndex"`
	OrganizationID string                `db:"organization_id" gorm:"type:varchar(36);not null;index"`
	UserID         string                `db:"user_id" gorm:"type:varchar(36);not null;index"`
	Scopes         string                `db:"scopes" gorm:"type:text"`
	ExpiresAt      *int64                `db:"expires_at" gorm:"type:bigint"`
	LastUsedAt     *int64                `db:"last_used_at" gorm:"type:bigint"`
	IsActive       bool                  `db:"is_active" gorm:"type:boolean;default:true"`
	CreatedAt      int64                 `db:"created_at" gorm:"type:bigint;autoCreateTime:milli"`
	UpdatedAt      int64                 `db:"updated_at" gorm:"type:bigint;autoUpdateTime:milli"`
	DeletedAt      soft_delete.DeletedAt `db:"deleted_at" gorm:"type:bigint;not null;default:0;softDelete:milli"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
