package worker_test

import (
	"testing"
	"time"

	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProcessorLifecycle(t *testing.T) {
	logger := logrus.New()
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}

	auditUC := new(auditMocks.MockAuditUseCase)
	auditRepo := new(auditMocks.MockAuditRepository)

	cleanupHandler := handlers.NewCleanupTaskHandler(nil, nil, nil, logger)
	webhookHandler := handlers.NewWebhookHandler(nil, logger)

	cfg := worker.WorkerConfig{
		SMTP: struct {
			Host       string
			Port       int
			Username   string
			Password   string
			FromSender string
			FromEmail  string
		}{Host: "smtp.example.com", Port: 587},
	}

	processor := worker.NewRedisTaskProcessor(redisOpt, logger, cleanupHandler, webhookHandler, auditUC, auditRepo, cfg)

	go func() {
		_ = processor.Start()
	}()
	time.Sleep(100 * time.Millisecond)
	processor.Shutdown()

	assert.NotNil(t, processor)
}

func TestAsynqLogger(t *testing.T) {
    logger := logrus.New()
    l := worker.NewAsynqLogger(logger)
    l.Info("info message")
    l.Debug("debug message")
    l.Warn("warn message")
    l.Error("error message")
    assert.NotNil(t, l)
}
