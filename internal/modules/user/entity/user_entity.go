package entity

import "gorm.io/plugin/soft_delete"

const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusBanned    = "banned"
)

type User struct {
	ID              string                `gorm:"column:id;primaryKey"`
	OrganizationID  *string               `gorm:"column:organization_id;index:idx_user_org_deleted;index"`
	Password        string                `gorm:"column:password"`
	Email           string                `gorm:"column:email;unique;not null"`
	Username        string                `gorm:"column:username;unique;not null"`
	Name            string                `gorm:"column:name"`
	AvatarURL       string                `gorm:"column:avatar_url"`
	Token           string                `gorm:"column:token"`
	Status          string                `gorm:"column:status;type:varchar(20);not null;default:'active';index:idx_user_status_deleted;index"`
	EmailVerifiedAt *int64                `gorm:"column:email_verified_at"`
	CreatedAt       int64                 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt       int64                 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt       soft_delete.DeletedAt `gorm:"column:deleted_at;softDelete:milli;index;index:idx_user_org_deleted;index:idx_user_status_deleted"`
	SSOIdentities   []UserSSOIdentity     `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type UserSSOIdentity struct {
	ID         string `gorm:"column:id;primaryKey;type:char(36)"`
	UserID     string `gorm:"column:user_id;type:char(36);not null;index"`
	Provider   string `gorm:"column:provider;type:varchar(50);not null"`
	ProviderID string `gorm:"column:provider_id;type:varchar(255);not null;uniqueIndex:idx_provider_id"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt  int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
