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
// This implements the "Global User, Local Member" identity model.
// A user can be a member of multiple organizations with different roles and statuses.
type OrganizationMember struct {
	ID             string `gorm:"column:id;primaryKey;type:varchar(36)"`
	OrganizationID string `gorm:"column:organization_id;type:varchar(36);not null;index"`
	UserID         string `gorm:"column:user_id;type:varchar(36);not null;index"`
	RoleID         string `gorm:"column:role_id;type:varchar(36);not null"`
	Status         string `gorm:"column:status;type:varchar(20);default:'active';index"`
	JoinedAt       int64  `gorm:"column:joined_at;autoCreateTime:milli"`

	// Relationships
	User userEntity.User `gorm:"foreignKey:UserID"`
	Role roleEntity.Role `gorm:"foreignKey:RoleID"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}
