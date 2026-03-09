package entity

import (
	"time"

	"gorm.io/gorm"
)

type ApiKey struct {
	ID             string         `gorm:"type:varchar(36);primaryKey"`
	Name           string         `gorm:"type:varchar(255);not null"`
	KeyHash        string         `gorm:"type:varchar(255);not null;uniqueIndex"`
	OrganizationID string         `gorm:"type:varchar(36);not null;index"`
	UserID         string         `gorm:"type:varchar(36);not null;index"`
	Scopes         string         `gorm:"type:text"`
	ExpiresAt      *time.Time     `gorm:"type:timestamp"`
	LastUsedAt     *time.Time     `gorm:"type:timestamp"`
	IsActive       bool           `gorm:"type:boolean;default:true"`
	CreatedAt      time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time      `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
