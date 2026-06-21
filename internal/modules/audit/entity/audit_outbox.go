package entity

const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusFailed     = "failed"
	OutboxStatusCompleted  = "completed"
)

type AuditOutbox struct {
	ID             string  `gorm:"primaryKey;type:varchar(36)"`
	OrganizationID *string `gorm:"type:varchar(36)"`
	UserID         string  `gorm:"type:varchar(36);not null"`
	Action         string  `gorm:"size:50;not null"`
	Entity         string  `gorm:"size:50;not null"`
	EntityID       string  `gorm:"size:100;not null"`
	OldValues      string  `gorm:"type:json"`
	NewValues      string  `gorm:"type:json"`
	IPAddress      string  `gorm:"size:45"`
	UserAgent      string  `gorm:"size:255"`
	Status         string  `gorm:"size:20;default:'pending'"`
	RetryCount     int     `gorm:"default:0"`
	LastError      string  `gorm:"type:text"`
	CreatedAt      int64   `gorm:"autoCreateTime:milli"`
	UpdatedAt      int64   `gorm:"autoCreateTime:milli;autoUpdateTime:milli"`
}

func (AuditOutbox) TableName() string {
	return "audit_outbox"
}
