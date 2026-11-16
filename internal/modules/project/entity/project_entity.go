package entity

type Project struct {
	ID             string `db:"id"`
	OrganizationID string `db:"organization_id"`
	UserID         string `db:"user_id"`
	Name           string `db:"name"`
	Domain         string `db:"domain"`
	Status         string `db:"status"`
	CreatedAt      int64  `db:"created_at"`
	UpdatedAt      int64  `db:"updated_at"`
	DeletedAt      int64  `db:"deleted_at"`
}

func (Project) TableName() string {
	return "projects"
}
