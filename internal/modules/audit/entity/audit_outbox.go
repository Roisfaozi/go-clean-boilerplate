package entity

const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusFailed     = "failed"
	OutboxStatusCompleted  = "completed"
)

type AuditOutbox struct {
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
	Status         string  `db:"status"`
	RetryCount     int     `db:"retry_count"`
	LastError      string  `db:"last_error"`
	CreatedAt      int64   `db:"created_at"`
	UpdatedAt      int64   `db:"updated_at"`
}

func (AuditOutbox) TableName() string {
	return "audit_outbox"
}
