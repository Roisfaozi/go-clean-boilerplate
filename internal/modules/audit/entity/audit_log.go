package entity

type AuditLog struct {
	ID             string  `db:"id"`
	OrganizationID *string `db:"organization_id"`
	UserID         string  `db:"user_id"`
	Action         string  `db:"action"`
	Entity         string  `db:"entity"`
	EntityID       string  `db:"entity_id"`
	OldValues      string  `db:"old_values"`
	NewValues      string  `db:"new_values"`
	IPAddress      string  `db:"ip_address"`
	UserAgent      string  `db:"user_agent"`
	CreatedAt      int64   `db:"created_at"`
	DeletedAt      int64   `db:"deleted_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
