package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync/atomic"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

func TestEmailTaskHandler_ProcessTaskSendEmail(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := handlers.SMTPConfig{
		Host:       "localhost",
		Port:       1025,
		Username:   "test",
		Password:   "test",
		FromSender: "Test",
		FromEmail:  "test@example.com",
	}
	handler := handlers.NewEmailTaskHandler(logger, cfg)

	t.Run("Valid Payload Processing", func(t *testing.T) {
		// This tests payload unmarshaling and handler logic
		// Real SMTP send will fail without a server, so we check for SMTP error
		payload := &tasks.SendEmailPayload{
			To:      "test@example.com",
			Subject: "Subject",
			Body:    "Body",
		}
		jsonPayload, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeSendEmail, jsonPayload)

		err := handler.ProcessTaskSendEmail(context.Background(), task)
		// SMTP will fail without a real server, but payload processing works
		if err != nil {
			assert.Contains(t, err.Error(), "failed to send email")
		}
	})

	t.Run("Unmarshal Error", func(t *testing.T) {
		task := asynq.NewTask(tasks.TypeSendEmail, []byte("invalid json"))

		err := handler.ProcessTaskSendEmail(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal task payload")
	})

	t.Run("MaskEmail Helper", func(t *testing.T) {
		assert.Equal(t, "al***@example.com", handlers.MaskEmail("alice@example.com"))
		assert.Equal(t, "***@domain.com", handlers.MaskEmail("a@domain.com"))
		assert.Equal(t, "***", handlers.MaskEmail("invalidemail"))
	})
}

func TestEmailTaskHandler_DedupeInMemory(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := handlers.SMTPConfig{
		Host:       "localhost",
		Port:       1025,
		FromSender: "Test",
		FromEmail:  "no-reply@example.com",
	}

	t.Run("suppresses duplicate by explicit idempotency key", func(t *testing.T) {
		handler := handlers.NewEmailTaskHandler(logger, cfg)
		var calls atomic.Int32
		handler.SetSendFunc(func(_ *gomail.Message) error {
			calls.Add(1)
			return nil
		})

		task1, err := tasks.NewSendEmailTaskWithPayload(&tasks.SendEmailPayload{
			To:             "user@example.com",
			Subject:        "Welcome",
			Body:           "<p>Hello</p>",
			IdempotencyKey: "welcome-user-1",
		})
		require.NoError(t, err)

		task2, err := tasks.NewSendEmailTaskWithPayload(&tasks.SendEmailPayload{
			To:             "user@example.com",
			Subject:        "Welcome (Retry)",
			Body:           "<p>Hello again</p>",
			IdempotencyKey: "welcome-user-1",
		})
		require.NoError(t, err)

		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task1))
		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task2))
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("suppresses duplicate by normalized content hash when idempotency key is empty", func(t *testing.T) {
		handler := handlers.NewEmailTaskHandler(logger, cfg)
		var calls atomic.Int32
		handler.SetSendFunc(func(_ *gomail.Message) error {
			calls.Add(1)
			return nil
		})

		task1, err := tasks.NewSendEmailTask("  Alice@Example.com ", "Reset", "<p>Link</p>")
		require.NoError(t, err)
		task2, err := tasks.NewSendEmailTask("alice@example.com", "Reset", "<p>Link</p>")
		require.NoError(t, err)
		task3, err := tasks.NewSendEmailTask("alice@example.com", "Reset", "<p>Different</p>")
		require.NoError(t, err)

		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task1))
		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task2))
		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task3))
		assert.Equal(t, int32(2), calls.Load())
	})

	t.Run("unmarks claim on SMTP failure so retry succeeds", func(t *testing.T) {
		handler := handlers.NewEmailTaskHandler(logger, cfg)
		var calls atomic.Int32
		handler.SetSendFunc(func(_ *gomail.Message) error {
			if calls.Add(1) == 1 {
				return errors.New("smtp connection refused")
			}
			return nil
		})

		task, err := tasks.NewSendEmailTaskWithPayload(&tasks.SendEmailPayload{
			To:             "retry@example.com",
			Subject:        "Retry Me",
			Body:           "<p>Body</p>",
			IdempotencyKey: "retry-key-1",
		})
		require.NoError(t, err)

		err = handler.ProcessTaskSendEmail(context.Background(), task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")

		// Retry should re-attempt and succeed.
		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task))
		assert.Equal(t, int32(2), calls.Load())

		// Replay after success must be suppressed.
		require.NoError(t, handler.ProcessTaskSendEmail(context.Background(), task))
		assert.Equal(t, int32(2), calls.Load())
	})
}

func TestEmailTaskHandler_DedupeRedis(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb.Close() }()

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := handlers.SMTPConfig{
		Host:       "localhost",
		Port:       1025,
		FromSender: "Test",
		FromEmail:  "no-reply@example.com",
	}

	t.Run("suppresses duplicate across two handler instances sharing Redis", func(t *testing.T) {
		handlerA := handlers.NewEmailTaskHandlerWithDedupe(logger, cfg, rdb)
		handlerB := handlers.NewEmailTaskHandler(logger, cfg)
		handlerB.SetRedisClient(rdb)

		var calls atomic.Int32
		send := func(_ *gomail.Message) error {
			calls.Add(1)
			return nil
		}
		handlerA.SetSendFunc(send)
		handlerB.SetSendFunc(send)

		task, err := tasks.NewSendEmailTaskWithPayload(&tasks.SendEmailPayload{
			To:             "shared@example.com",
			Subject:        "Invite",
			Body:           "<p>Join</p>",
			IdempotencyKey: "invite-org-1-user-9",
		})
		require.NoError(t, err)

		require.NoError(t, handlerA.ProcessTaskSendEmail(context.Background(), task))
		require.NoError(t, handlerB.ProcessTaskSendEmail(context.Background(), task))
		assert.Equal(t, int32(1), calls.Load())
		assert.True(t, mr.Exists("worker:email:dedupe:invite-org-1-user-9"))
	})

	t.Run("unmarks Redis claim on SMTP error allowing retry", func(t *testing.T) {
		handlerA := handlers.NewEmailTaskHandlerWithDedupe(logger, cfg, rdb)
		handlerB := handlers.NewEmailTaskHandlerWithDedupe(logger, cfg, rdb)

		var calls atomic.Int32
		handlerA.SetSendFunc(func(_ *gomail.Message) error {
			calls.Add(1)
			return errors.New("temporary smtp timeout")
		})
		handlerB.SetSendFunc(func(_ *gomail.Message) error {
			calls.Add(1)
			return nil
		})

		task, err := tasks.NewSendEmailTaskWithPayload(&tasks.SendEmailPayload{
			To:             "retry-redis@example.com",
			Subject:        "Verify",
			Body:           "<p>Verify</p>",
			IdempotencyKey: "verify-user-42",
		})
		require.NoError(t, err)

		err = handlerA.ProcessTaskSendEmail(context.Background(), task)
		require.Error(t, err)
		assert.False(t, mr.Exists("worker:email:dedupe:verify-user-42"))

		require.NoError(t, handlerB.ProcessTaskSendEmail(context.Background(), task))
		assert.Equal(t, int32(2), calls.Load())
		assert.True(t, mr.Exists("worker:email:dedupe:verify-user-42"))
	})
}
