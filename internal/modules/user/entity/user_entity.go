package entity

import "gorm.io/plugin/soft_delete"

const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusBanned    = "banned"
)

type User struct {
	ID        string                `gorm:"column:id;primaryKey"`
	Password  string                `gorm:"column:password"`
	Email     string                `gorm:"column:email;unique;not null"`
	Username  string                `gorm:"column:username;unique;not null"`
	Name      string                `gorm:"column:name"`
	AvatarURL string                `gorm:"column:avatar_url"`
	Token     string                `gorm:"column:token"`
	Status          string                `gorm:"column:status;type:varchar(20);not null;default:'active';index"`
	EmailVerifiedAt *int64                `gorm:"column:email_verified_at"`
	CreatedAt       int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index"`
}

