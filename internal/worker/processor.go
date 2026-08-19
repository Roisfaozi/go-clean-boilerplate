package worker

import (
	"context"
	"sync"
	"time"

	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
}

type RedisTaskProcessor struct {
	server         *asynq.Server
	logger         *logrus.Logger
	cleanupHandler *handlers.CleanupTaskHandler
	webhookHandler *handlers.WebhookHandler
	auditUC        auditUseCase.AuditUseCase
	auditRepo      auditUseCase.AuditRepository
	cfg            WorkerConfig
	mu             sync.Mutex
	started        bool
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	logger *logrus.Logger,
	cleanupHandler *handlers.CleanupTaskHandler,
	webhookHandler *handlers.WebhookHandler,
	auditUC auditUseCase.AuditUseCase,
	auditRepo auditUseCase.AuditRepository,
	cfg WorkerConfig,
) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				queueNameCritical: 6,
				queueNameDefault:  3,
				queueNameLow:      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.WithContext(ctx).Errorf("Failed to process task type %s: %v", task.Type(), err)
			}),
			Logger: NewAsynqLogger(logger),
		},
	)

	return &RedisTaskProcessor{
		server:         server,
		logger:         logger,
		cleanupHandler: cleanupHandler,
		webhookHandler: webhookHandler,
		auditUC:        auditUC,
		auditRepo:      auditRepo,
		cfg:            cfg,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	processor.mu.Lock()
	if processor.started {
		processor.mu.Unlock()
		return nil
	}
	processor.started = true
	processor.mu.Unlock()

	mux := asynq.NewServeMux()

	// Map WorkerConfig to Handler Config
	smtpCfg := handlers.SMTPConfig{
		Host:       processor.cfg.SMTP.Host,
		Port:       processor.cfg.SMTP.Port,
		Username:   processor.cfg.SMTP.Username,
		Password:   processor.cfg.SMTP.Password,
		FromSender: processor.cfg.SMTP.FromSender,
		FromEmail:  processor.cfg.SMTP.FromEmail,
	}

	registerInstrumented := func(taskType string, fn func(context.Context, *asynq.Task) error) {
		mux.HandleFunc(taskType, InstrumentTaskHandler(taskType, fn))
	}

	emailHandler := handlers.NewEmailTaskHandler(processor.logger, smtpCfg)
	registerInstrumented(tasks.TypeSendEmail, emailHandler.ProcessTaskSendEmail)

	auditHandler := handlers.NewAuditTaskHandler(processor.logger, processor.auditUC)
	registerInstrumented(tasks.TypeAuditLogCreate, auditHandler.ProcessTaskAuditLog)
	registerInstrumented(tasks.TypeAuditLogExport, auditHandler.ProcessTaskAuditLogExport)

	outboxHandler := handlers.NewOutboxTaskHandler(processor.auditRepo, processor.logger)
	registerInstrumented(tasks.TypeAuditOutboxSync, outboxHandler.ProcessAuditOutbox)

	if processor.webhookHandler != nil {
		registerInstrumented(tasks.TypeWebhookTrigger, processor.webhookHandler.ProcessTaskWebhookTrigger)
	}

	// Register Cleanup Handlers
	if processor.cleanupHandler != nil {
		registerInstrumented(tasks.TypeCleanupExpiredTokens, processor.cleanupHandler.ProcessCleanupExpiredTokens)
		registerInstrumented(tasks.TypeCleanupSoftDeletedEntities, processor.cleanupHandler.ProcessCleanupSoftDeletedEntities)
		registerInstrumented(tasks.TypePruneAuditLogs, processor.cleanupHandler.ProcessPruneAuditLogs)
	}

	if err := processor.server.Start(mux); err != nil {
		processor.mu.Lock()
		processor.started = false
		processor.mu.Unlock()
		return err
	}

	return nil
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.mu.Lock()
	if !processor.started {
		processor.mu.Unlock()
		return
	}
	processor.started = false
	processor.mu.Unlock()

	processor.server.Shutdown()
}

func InstrumentTaskHandler(taskType string, fn func(context.Context, *asynq.Task) error) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		start := time.Now()
		err := fn(ctx, task)
		duration := time.Since(start).Seconds()

		status := "success"
		if err != nil {
			status = "failed"
		}

		telemetry.WorkerTasksTotal.WithLabelValues(taskType, status).Inc()
		telemetry.WorkerTaskDuration.WithLabelValues(taskType, status).Observe(duration)

		return err
	}
}

type AsynqLogger struct {
	logger *logrus.Logger
}

func NewAsynqLogger(logger *logrus.Logger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

func (l *AsynqLogger) Debug(args ...interface{}) { l.logger.Debug(args...) }
func (l *AsynqLogger) Info(args ...interface{})  { l.logger.Info(args...) }
func (l *AsynqLogger) Warn(args ...interface{})  { l.logger.Warn(args...) }
func (l *AsynqLogger) Error(args ...interface{}) { l.logger.Error(args...) }
func (l *AsynqLogger) Fatal(args ...interface{}) { l.logger.Fatal(args...) }
