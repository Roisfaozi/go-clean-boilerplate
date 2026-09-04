package entity

// InvitationToken represents an invitation for a user to join an organization
type InvitationToken struct {
	ID             string `db:"id" gorm:"primaryKey;type:varchar(36)"`
	OrganizationID string `db:"organization_id" gorm:"type:varchar(36);not null;index"`
	Email          string `db:"email" gorm:"type:varchar(255);not null;index"`
	Token          string `db:"token" gorm:"type:varchar(255);unique;not null;index"`
	RoleID         string `db:"role_id" gorm:"column:role_id;type:varchar(36);not null"`
	ExpiresAt      int64  `db:"expires_at" gorm:"type:bigint;not null"`
	CreatedAt      int64  `db:"created_at" gorm:"type:bigint;not null"`
}

// TableName specifies the table name for InvitationToken
func (InvitationToken) TableName() string {
	return "invitation_tokens"
}
