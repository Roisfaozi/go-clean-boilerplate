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
	ID             string `db:"id"`
	OrganizationID string `db:"organization_id"`
	UserID         string `db:"user_id"`
	RoleID         string `db:"role_id"`
	Status         string `db:"status"`
	JoinedAt       int64  `db:"joined_at"`

	// Relationships
	User userEntity.User `db:"-"`
	Role roleEntity.Role `db:"-"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}
