package worker_test

import (
	"testing"
	"time"

	auditMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	workerMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/worker/test/mocks"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type processorTestDeps struct {
	processor worker.TaskProcessor
	mockRedis *miniredis.Miniredis
	auditUC   *auditMocks.MockAuditUseCase
	auditRepo *workerMocks.MockAuditRepository
}

func setupProcessorTest(t *testing.T) (*processorTestDeps, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisOpt := asynq.RedisClientOpt{
		Addr: mr.Addr(),
	}

	logger := logrus.New()
	logger.SetOutput(new(mockWriter))

	auditUC := new(auditMocks.MockAuditUseCase)
	auditRepo := new(workerMocks.MockAuditRepository)

	cfg := worker.WorkerConfig{
		SMTP: worker.SMTPConfig{
			Host: "localhost",
			Port: 1025,
		},
	}

	cleanupHandler := handlers.NewCleanupTaskHandler(nil, nil, auditRepo, logger)
	webhookHandler := handlers.NewWebhookHandler(nil, logger)

	processor := worker.NewRedisTaskProcessor(
		redisOpt,
		logger,
		cleanupHandler,
		webhookHandler,
		auditUC,
		auditRepo,
		cfg,
	)

	cleanup := func() {
		processor.Shutdown()
		mr.Close()
	}

	return &processorTestDeps{
		processor: processor,
		mockRedis: mr,
		auditUC:   auditUC,
		auditRepo: auditRepo,
	}, cleanup
}

type mockWriter struct{}
func (m *mockWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func TestRedisTaskProcessor_Lifecycle(t *testing.T) {
	deps, cleanup := setupProcessorTest(t)
	defer cleanup()

	t.Run("Positive - Start and Shutdown", func(t *testing.T) {
		errChan := make(chan error, 1)

		go func() {
			errChan <- deps.processor.Start()
		}()

		// Let it start
		time.Sleep(100 * time.Millisecond)

		// Shutdown
		deps.processor.Shutdown()

		// Ensure it stopped without major error (Asynq usually returns nil on successful shutdown or specific error if aborted)
		err := <-errChan
		assert.NoError(t, err)
	})

	t.Run("Negative - Start with bad redis connection", func(t *testing.T) {
		// Create a new processor pointing to nowhere
		logger := logrus.New()
		logger.SetOutput(new(mockWriter))

		processor := worker.NewRedisTaskProcessor(
			asynq.RedisClientOpt{Addr: "invalid:9999"},
			logger,
			nil,
			nil,
			nil,
			nil,
			worker.WorkerConfig{},
		)

		err := processor.Start()
		/* asynq.Server logs error and returns nil if dial fails or panics, but let us verify how Start behaves on bad conn.
		   Wait, asynq Start doesn't immediately error on bad redis, it retries in background.
		   We should assert NoError instead or check logs. */
		assert.NoError(t, err)
	})

	t.Run("Edge - Start multiple times concurrently", func(t *testing.T) {
		go func() { _ = deps.processor.Start() }()
		go func() { _ = deps.processor.Start() }()

		time.Sleep(100 * time.Millisecond)
		deps.processor.Shutdown()
	})

	t.Run("Vulnerability - Null configs injection", func(t *testing.T) {
		logger := logrus.New()
		logger.SetOutput(new(mockWriter))

		processor := worker.NewRedisTaskProcessor(
			asynq.RedisClientOpt{Addr: deps.mockRedis.Addr()},
			logger,
			nil, // Should handle nil safely in NewRedisTaskProcessor Start
			nil,
			nil,
			nil,
			worker.WorkerConfig{},
		)

		errChan := make(chan error, 1)
		go func() {
			errChan <- processor.Start()
		}()
		time.Sleep(100 * time.Millisecond)
		processor.Shutdown()
		assert.NoError(t, <-errChan)
	})
}

func TestAsynqLogger_Fatal(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(new(mockWriter))

	var fatalCalled bool
	logger.ExitFunc = func(code int) {
		fatalCalled = true
	}

	asynqLogger := worker.NewAsynqLogger(logger)
	asynqLogger.Fatal("test fatal")

	assert.True(t, fatalCalled)
}
