package worker

import (
	"encoding/json"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	scheduler *asynq.Scheduler
	logger    *logrus.Logger
}

func NewScheduler(redisOpt asynq.RedisClientOpt, logger *logrus.Logger) *Scheduler {
	location, _ := time.LoadLocation("Asia/Jakarta") // Adjust timezone as needed

	scheduler := asynq.NewScheduler(
		redisOpt,
		&asynq.SchedulerOpts{
			Location: location,
			Logger:   NewAsynqLogger(logger),
		},
	)

	return &Scheduler{
		scheduler: scheduler,
		logger:    logger,
	}
}

// RegisterScheduledTasks registers all periodic tasks
func (s *Scheduler) RegisterScheduledTasks() {
	// 1. Cleanup Expired Reset Tokens
	// Run every 6 hours
	if _, err := s.scheduler.Register("@every 6h", asynq.NewTask(tasks.TypeCleanupExpiredTokens, nil)); err != nil {
		s.logger.Errorf("Failed to register task %s: %v", tasks.TypeCleanupExpiredTokens, err)
	}

	// 2. Hard Delete Soft-Deleted Entities (Users)
	// Run daily at 03:00 AM
	// Retention: 30 days
	payloadUser, _ := json.Marshal(tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 30})
	if _, err := s.scheduler.Register("0 3 * * *", asynq.NewTask(tasks.TypeCleanupSoftDeletedEntities, payloadUser)); err != nil {
		s.logger.Errorf("Failed to register task %s: %v", tasks.TypeCleanupSoftDeletedEntities, err)
	}

	// 3. Prune Old Audit Logs
	// Run weekly (Sunday at 04:00 AM)
	// Retention: 180 days (6 months)
	payloadAudit, _ := json.Marshal(tasks.PruneAuditLogsPayload{RetentionDays: 180})
	if _, err := s.scheduler.Register("0 4 * * 0", asynq.NewTask(tasks.TypePruneAuditLogs, payloadAudit)); err != nil {
		s.logger.Errorf("Failed to register task %s: %v", tasks.TypePruneAuditLogs, err)
	}

	// 4. Audit Outbox Sync
	// Run every 30 seconds
	if _, err := s.scheduler.Register("@every 30s", tasks.NewAuditOutboxSyncTask()); err != nil {
		s.logger.Errorf("Failed to register task %s: %v", tasks.TypeAuditOutboxSync, err)
	}

	s.logger.Info("Scheduled tasks registered successfully")
}

// Start starts the scheduler. This is a blocking call, so run it in a goroutine.
func (s *Scheduler) Start() error {
	return s.scheduler.Run()
}

// Shutdown stops the scheduler
func (s *Scheduler) Shutdown() {
	s.scheduler.Shutdown()
}
