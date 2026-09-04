package entity

import (
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
)

const (
	MemberStatusActive    = "active"
	MemberStatusInvited   = "invited"
	MemberStatusSuspended = "suspended"
	MemberStatusBanned    = "banned"
)

// OrganizationMember represents the membership pivot between Users and Organizations.
type OrganizationMember struct {
	ID             string `db:"id" gorm:"column:id;primaryKey;type:varchar(36)"`
	OrganizationID string `db:"organization_id" gorm:"column:organization_id;type:varchar(36);not null;index"`
	UserID         string `db:"user_id" gorm:"column:user_id;type:varchar(36);not null;index"`
	RoleID         string `db:"role_id" gorm:"column:role_id;type:varchar(36);not null"`
	Status         string `db:"status" gorm:"column:status;type:varchar(20);default:'active';index"`
	JoinedAt       int64  `db:"joined_at" gorm:"column:joined_at;autoCreateTime:milli"`

	// Relationships
	User userEntity.User `db:"-" gorm:"foreignKey:UserID"`
	Role roleEntity.Role `db:"-" gorm:"foreignKey:RoleID"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}
