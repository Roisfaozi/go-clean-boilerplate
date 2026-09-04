package entity

import "gorm.io/plugin/soft_delete"

const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusBanned    = "banned"
)

type User struct {
	ID              string                `db:"id" gorm:"column:id;primaryKey"`
	OrganizationID  *string               `db:"organization_id" gorm:"column:organization_id;index:idx_user_org_deleted;index"`
	Password        string                `db:"password" gorm:"column:password"`
	Email           string                `db:"email" gorm:"column:email;unique;not null"`
	Username        string                `db:"username" gorm:"column:username;unique;not null"`
	Name            string                `db:"name" gorm:"column:name"`
	AvatarURL       string                `db:"avatar_url" gorm:"column:avatar_url"`
	Token           string                `db:"token" gorm:"column:token"`
	Status          string                `db:"status" gorm:"column:status;type:varchar(20);not null;default:'active';index:idx_user_status_deleted;index"`
	EmailVerifiedAt *int64                `db:"email_verified_at" gorm:"column:email_verified_at"`
	CreatedAt       int64                 `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt       int64                 `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
	DeletedAt       soft_delete.DeletedAt `db:"deleted_at" gorm:"column:deleted_at;softDelete:milli;index;index:idx_user_org_deleted;index:idx_user_status_deleted"`
	SSOIdentities   []UserSSOIdentity     `db:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}

type UserSSOIdentity struct {
	ID         string `db:"id" gorm:"column:id;primaryKey;type:char(36)"`
	UserID     string `db:"user_id" gorm:"column:user_id;type:char(36);not null;index"`
	Provider   string `db:"provider" gorm:"column:provider;type:varchar(50);not null"`
	ProviderID string `db:"provider_id" gorm:"column:provider_id;type:varchar(255);not null;uniqueIndex:idx_provider_id"`
	CreatedAt  int64  `db:"created_at" gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt  int64  `db:"updated_at" gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (UserSSOIdentity) TableName() string {
	return "user_sso_identities"
}
