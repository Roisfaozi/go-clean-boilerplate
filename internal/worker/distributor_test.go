package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type distributorTestDeps struct {
	distributor worker.TaskDistributor
	mockRedis   *miniredis.Miniredis
}

func setupDistributorTest(t *testing.T) (*distributorTestDeps, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisOpt := asynq.RedisClientOpt{
		Addr: mr.Addr(),
	}

	distributor := worker.NewRedisTaskDistributor(redisOpt)

	cleanup := func() {
		if redisDist, ok := distributor.(*worker.RedisTaskDistributor); ok {
			_ = redisDist.Close()
		}
		mr.Close()
	}

	return &distributorTestDeps{
		distributor: distributor,
		mockRedis:   mr,
	}, cleanup
}

func TestRedisTaskDistributor_DistributeTaskWebhookTrigger(t *testing.T) {
	t.Run("Positive - Enqueue Webhook Trigger", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := tasks.WebhookTriggerPayload{
			EventType: "user.created",
			Payload:   `{"id": "123"}`,
		}

		err := deps.distributor.DistributeTaskWebhookTrigger(context.Background(), payload, asynq.MaxRetry(3))
		assert.NoError(t, err)
	})

	t.Run("Negative - Closed Connection", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		if dist, ok := deps.distributor.(*worker.RedisTaskDistributor); ok {
			_ = dist.Close()
		}

		payload := tasks.WebhookTriggerPayload{
			EventType: "user.deleted",
		}

		err := deps.distributor.DistributeTaskWebhookTrigger(context.Background(), payload)
		assert.Error(t, err)
	})

	t.Run("Edge - Empty Payload", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := tasks.WebhookTriggerPayload{}
		err := deps.distributor.DistributeTaskWebhookTrigger(context.Background(), payload)
		assert.NoError(t, err)
	})

	t.Run("Vulnerability - Malicious Event String", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := tasks.WebhookTriggerPayload{
			EventType: "<script>alert(1)</script>",
		}
		err := deps.distributor.DistributeTaskWebhookTrigger(context.Background(), payload)
		assert.NoError(t, err)
	})
}

func TestRedisTaskDistributor_DistributeTaskSendEmail(t *testing.T) {
	t.Run("Positive - Enqueue Email Task", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := &tasks.SendEmailPayload{
			To:      "test@example.com",
			Subject: "Welcome",
			Body:    "<h1>Welcome!</h1>",
		}

		err := deps.distributor.DistributeTaskSendEmail(context.Background(), payload)
		assert.NoError(t, err)
	})

	t.Run("Negative - Closed Connection Error", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		if dist, ok := deps.distributor.(*worker.RedisTaskDistributor); ok {
			_ = dist.Close()
		}

		err := deps.distributor.DistributeTaskSendEmail(context.Background(), &tasks.SendEmailPayload{})
		assert.Error(t, err)
	})

	t.Run("Edge - Extremely Long Subject", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		longSubject := string(make([]byte, 10000))
		payload := &tasks.SendEmailPayload{
			To:      "test@example.com",
			Subject: longSubject,
			Body:    "Body",
		}

		err := deps.distributor.DistributeTaskSendEmail(context.Background(), payload)
		assert.NoError(t, err)
	})

	t.Run("Vulnerability - XSS in Email Body", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := &tasks.SendEmailPayload{
			To:      "test@example.com",
			Subject: "Subject",
			Body:    "<script>alert(1)</script><img src=x onerror=alert(1)>",
		}

		err := deps.distributor.DistributeTaskSendEmail(context.Background(), payload)
		assert.NoError(t, err)
	})
}

func TestRedisTaskDistributor_DistributeTaskAuditLog(t *testing.T) {
	t.Run("Positive - Enqueue Audit Log", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := model.CreateAuditLogRequest{
			UserID: "user_123",
			Action: "LOGIN",
		}

		err := deps.distributor.DistributeTaskAuditLog(context.Background(), payload)
		assert.NoError(t, err)
	})

	t.Run("Negative - Connection Failure", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		if dist, ok := deps.distributor.(*worker.RedisTaskDistributor); ok {
			_ = dist.Close()
		}

		err := deps.distributor.DistributeTaskAuditLog(context.Background(), model.CreateAuditLogRequest{})
		assert.Error(t, err)
	})

	t.Run("Edge - Large Old/New Values", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := model.CreateAuditLogRequest{
			UserID: "user_123",
			OldValues: map[string]interface{}{
				"data": string(make([]byte, 5000)),
			},
		}

		err := deps.distributor.DistributeTaskAuditLog(context.Background(), payload)
		assert.NoError(t, err)
	})
}

func TestRedisTaskDistributor_DistributeTaskAuditOutboxSync(t *testing.T) {
	t.Run("Positive - Enqueue Outbox Sync", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		err := deps.distributor.DistributeTaskAuditOutboxSync(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Negative - Connection Failure", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		if dist, ok := deps.distributor.(*worker.RedisTaskDistributor); ok {
			_ = dist.Close()
		}

		err := deps.distributor.DistributeTaskAuditOutboxSync(context.Background())
		assert.Error(t, err)
	})
}

func TestRedisTaskDistributor_DistributeTaskAuditLogExport(t *testing.T) {
	t.Run("Positive - Enqueue Export", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := model.AuditLogExportPayload{
			UserID:   "user_123",
			FromDate: time.Now().Add(-24*time.Hour).Format("2006-01-02"),
			ToDate:   time.Now().Format("2006-01-02"),
		}

		err := deps.distributor.DistributeTaskAuditLogExport(context.Background(), payload)
		assert.NoError(t, err)
	})

	t.Run("Negative - Connection Failure", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		if dist, ok := deps.distributor.(*worker.RedisTaskDistributor); ok {
			_ = dist.Close()
		}

		err := deps.distributor.DistributeTaskAuditLogExport(context.Background(), model.AuditLogExportPayload{})
		assert.Error(t, err)
	})

	t.Run("Vulnerability - SQLi format injection attempt", func(t *testing.T) {
		deps, cleanup := setupDistributorTest(t)
		defer cleanup()
		payload := model.AuditLogExportPayload{
			UserID: "user_123'; DROP TABLE users;--",
		}

		err := deps.distributor.DistributeTaskAuditLogExport(context.Background(), payload)
		assert.NoError(t, err)
	})
}
