package entity

const (
	OrgStatusActive    = "active"
	OrgStatusSuspended = "suspended"
	OrgStatusInactive  = "inactive"
)

// Organization represents a tenant/workspace in the multi-tenant system.
// Users can belong to multiple organizations with different roles.
type Organization struct {
	ID        string                 `db:"id"`
	Name      string                 `db:"name"`
	Slug      string                 `db:"slug"`
	OwnerID   string                 `db:"owner_id"`
	Settings  map[string]interface{} `db:"-"`
	Status    string                 `db:"status"`
	CreatedAt int64                  `db:"created_at"`
	UpdatedAt int64                  `db:"updated_at"`
	DeletedAt int64                  `db:"deleted_at"`

	// Relations
	Members []OrganizationMember `db:"-"`
}

func (Organization) TableName() string {
	return "organizations"
}
