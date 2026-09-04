package entity

const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusFailed     = "failed"
	OutboxStatusCompleted  = "completed"
)

type AuditOutbox struct {
	ID             string  `db:"id" gorm:"primaryKey;type:varchar(36)"`
	OrganizationID *string `db:"organization_id" gorm:"type:varchar(36)"`
	UserID         string  `db:"user_id" gorm:"type:varchar(36);not null"`
	Action         string  `db:"action" gorm:"size:50;not null"`
	Entity         string  `db:"entity" gorm:"size:50;not null"`
	EntityID       string  `db:"entity_id" gorm:"size:100;not null"`
	OldValues      string  `db:"old_values" gorm:"type:json"`
	NewValues      string  `db:"new_values" gorm:"type:json"`
	IPAddress      string  `db:"ip_address" gorm:"size:45"`
	UserAgent      string  `db:"user_agent" gorm:"size:255"`
	Status         string  `db:"status" gorm:"size:20;default:'pending'"`
	RetryCount     int     `db:"retry_count" gorm:"default:0"`
	LastError      string  `db:"last_error" gorm:"type:text"`
	CreatedAt      int64   `db:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt      int64   `db:"updated_at" gorm:"autoCreateTime:milli;autoUpdateTime:milli"`
}

func (AuditOutbox) TableName() string {
	return "audit_outbox"
}
