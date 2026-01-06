package entity

import "gorm.io/gorm"

const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusBanned    = "banned"
)

type User struct {
	ID        string         `gorm:"column:id;primaryKey"`
	Password  string         `gorm:"column:password"`
	Email     string         `gorm:"column:email;unique;not null"`
	Username  string         `gorm:"column:username;unique;not null"`
	Name      string         `gorm:"column:name"`
	Token     string         `gorm:"column:token"`
	Status    string         `gorm:"column:status;type:varchar(20);not null;default:'active';index"`
	CreatedAt int64          `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64          `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}
