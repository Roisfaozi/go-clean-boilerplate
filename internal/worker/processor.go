package worker

import (
	"context"

	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
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
	auditUC        auditUseCase.AuditUseCase
	auditRepo      auditUseCase.AuditRepository
	cfg            WorkerConfig
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	logger *logrus.Logger,
	cleanupHandler *handlers.CleanupTaskHandler,
	auditUC auditUseCase.AuditUseCase,
	auditRepo auditUseCase.AuditRepository,
	cfg WorkerConfig,
) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
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
		auditUC:        auditUC,
		auditRepo:      auditRepo,
		cfg:            cfg,
	}
}

func (processor *RedisTaskProcessor) Start() error {
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

	emailHandler := handlers.NewEmailTaskHandler(processor.logger, smtpCfg)
	mux.HandleFunc(tasks.TypeSendEmail, emailHandler.ProcessTaskSendEmail)

	auditHandler := handlers.NewAuditTaskHandler(processor.logger, processor.auditUC)
	mux.HandleFunc(tasks.TypeAuditLogCreate, auditHandler.ProcessTaskAuditLog)

	outboxHandler := handlers.NewOutboxTaskHandler(processor.auditRepo, processor.logger)
	mux.HandleFunc(tasks.TypeAuditOutboxSync, outboxHandler.ProcessAuditOutbox)

	// Register Cleanup Handlers
	if processor.cleanupHandler != nil {
		mux.HandleFunc(tasks.TypeCleanupExpiredTokens, processor.cleanupHandler.ProcessCleanupExpiredTokens)
		mux.HandleFunc(tasks.TypeCleanupSoftDeletedEntities, processor.cleanupHandler.ProcessCleanupSoftDeletedEntities)
		mux.HandleFunc(tasks.TypePruneAuditLogs, processor.cleanupHandler.ProcessPruneAuditLogs)
	}

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
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
