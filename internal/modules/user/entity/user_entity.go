package entity

const (
	UserStatusActive    = "active"
	UserStatusSuspended = "suspended"
	UserStatusBanned    = "banned"
)

type User struct {
	ID              string            `db:"id"`
	OrganizationID  *string           `db:"organization_id"`
	Password        string            `db:"password"`
	Email           string            `db:"email"`
	Username        string            `db:"username"`
	Name            string            `db:"name"`
	AvatarURL       string            `db:"avatar_url"`
	Token           string            `db:"token"`
	Status          string            `db:"status"`
	EmailVerifiedAt *int64            `db:"email_verified_at"`
	CreatedAt       int64             `db:"created_at"`
	UpdatedAt       int64             `db:"updated_at"`
	DeletedAt       int64             `db:"deleted_at"`
	SSOIdentities   []UserSSOIdentity `db:"-"`
}

func (User) TableName() string {
	return "users"
}

type UserSSOIdentity struct {
	ID         string `db:"id"`
	UserID     string `db:"user_id"`
	Provider   string `db:"provider"`
	ProviderID string `db:"provider_id"`
	CreatedAt  int64  `db:"created_at"`
	UpdatedAt  int64  `db:"updated_at"`
}

func (UserSSOIdentity) TableName() string {
	return "user_sso_identities"
}
