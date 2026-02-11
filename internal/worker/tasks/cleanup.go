package tasks

// Cleanup Task Types
const (
	TypeCleanupExpiredTokens       = "cleanup:expired_tokens"
	TypeCleanupSoftDeletedEntities = "cleanup:soft_deleted_entities"
	TypePruneAuditLogs             = "cleanup:prune_audit_logs"
	TypePrunePresence              = "cleanup:prune_presence"
)

// Payload for CleanupSoftDeletedEntities
type CleanupSoftDeletedEntitiesPayload struct {
	RetentionDays int `json:"retention_days"`
}

// Payload for PruneAuditLogs
type PruneAuditLogsPayload struct {
	RetentionDays int `json:"retention_days"`
}
