package entity

// InvitationToken represents an invitation for a user to join an organization
type InvitationToken struct {
	ID             string `gorm:"primaryKey;type:varchar(36)"`
	OrganizationID string `gorm:"type:varchar(36);not null;index"`
	Email          string `gorm:"type:varchar(255);not null;index"`
	Token          string `gorm:"type:varchar(255);unique;not null;index"`
	Role           string `gorm:"type:varchar(36);not null"`
	ExpiresAt      int64  `gorm:"type:bigint;not null"`
	CreatedAt      int64  `gorm:"type:bigint;not null"`
}

// TableName specifies the table name for InvitationToken
func (InvitationToken) TableName() string {
	return "invitation_tokens"
}
