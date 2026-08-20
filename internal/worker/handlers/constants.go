package handlers

const (
	cleanupTaskLabelExpiredTokens       = "expired_tokens"
	cleanupTaskLabelSoftDeletedEntities = "soft_deleted_entities"
	cleanupTaskLabelPruneAuditLogs      = "prune_audit_logs"
	cleanupTaskStatusFailed             = "failed"
	cleanupTaskStatusSuccess            = "success"
)

const (
	headerContentType          = "Content-Type"
	headerValueApplicationJSON = "application/json"
	headerWebhookSignature     = "X-Webhook-Signature"
	headerWebhookEvent         = "X-Webhook-Event"
	headerWebhookID            = "X-Webhook-ID"
	headerWebhookTimestamp     = "X-Webhook-Timestamp"
	headerEmailFrom            = "From"
	headerEmailTo              = "To"
	headerEmailSubject         = "Subject"
	headerValueTextHTML        = "text/html"
)
