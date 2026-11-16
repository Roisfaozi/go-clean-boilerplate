package entity

type Role struct {
	ID             string  `db:"id"`
	Name           string  `db:"name"`
	OrganizationID *string `db:"organization_id"`
	Description    string  `db:"description"`
	CreatedAt      int64   `db:"created_at"`
	UpdatedAt      int64   `db:"updated_at"`
	DeletedAt      int64   `db:"deleted_at"`
}

func (Role) TableName() string {
	return "roles"
}
