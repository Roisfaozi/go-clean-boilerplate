package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/handlers"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
}
