package entity

import "gorm.io/gorm"

type Role struct {
	ID          string         `gorm:"type:varchar(36);primary_key"`
	Name        string         `gorm:"type:varchar(50);not null;unique"`
	Description string         `gorm:"type:text"`
	CreatedAt   int64          `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   int64          `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
