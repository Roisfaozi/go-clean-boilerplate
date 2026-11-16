package entity

type AccessRight struct {
	ID             string     `db:"id"`
	OrganizationID *string    `db:"organization_id"`
	Name           string     `db:"name"`
	Description    string     `db:"description"`
	Endpoints      []Endpoint `db:"-"`
	CreatedAt      int64      `db:"created_at"`
	UpdatedAt      int64      `db:"updated_at"`
	DeletedAt      int64      `db:"deleted_at"`
}

func (AccessRight) TableName() string {
	return "access_rights"
}

type Endpoint struct {
	ID        string `db:"id"`
	Path      string `db:"path"`
	Method    string `db:"method"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
	DeletedAt int64  `db:"deleted_at"`
}

func (Endpoint) TableName() string {
	return "endpoints"
}

type AccessRightEndpoint struct {
	AccessRightID string `db:"access_right_id"`
	EndpointID    string `db:"endpoint_id"`
}

func (AccessRightEndpoint) TableName() string {
	return "access_right_endpoints"
}
