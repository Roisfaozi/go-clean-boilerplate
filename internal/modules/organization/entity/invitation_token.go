package entity

// InvitationToken represents an invitation for a user to join an organization
type InvitationToken struct {
	ID             string `db:"id"`
	OrganizationID string `db:"organization_id"`
	Email          string `db:"email"`
	Token          string `db:"token"`
	RoleID         string `db:"role_id"`
	ExpiresAt      int64  `db:"expires_at"`
	CreatedAt      int64  `db:"created_at"`
}

// TableName specifies the table name for InvitationToken
func (InvitationToken) TableName() string {
	return "invitation_tokens"
}
