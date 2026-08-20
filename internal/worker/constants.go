package worker

const (
	queueNameCritical           = "critical"
	queueNameDefault            = "default"
	queueNameLow                = "low"
	schedulerLocationName       = "Asia/Jakarta"
	cleanupScheduleEvery6h      = "@every 6h"
	cleanupScheduleDaily3am     = "0 3 * * *"
	pruneAuditScheduleWeekly4am = "0 4 * * 0"
	outboxSyncScheduleEvery5s   = "@every 5s"
)
