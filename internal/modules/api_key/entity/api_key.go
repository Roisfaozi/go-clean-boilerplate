package entity

type ApiKey struct {
	ID             string `db:"id"`
	Name           string `db:"name"`
	KeyHash        string `db:"key_hash"`
	OrganizationID string `db:"organization_id"`
	UserID         string `db:"user_id"`
	Scopes         string `db:"scopes"`
	ExpiresAt      *int64 `db:"expires_at"`
	LastUsedAt     *int64 `db:"last_used_at"`
	IsActive       bool   `db:"is_active"`
	CreatedAt      int64  `db:"created_at"`
	UpdatedAt      int64  `db:"updated_at"`
	DeletedAt      int64  `db:"deleted_at"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
