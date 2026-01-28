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
	handler := handlers.NewEmailTaskHandler(logger)

	t.Run("Success", func(t *testing.T) {
		payload := &tasks.SendEmailPayload{
			To:      "test@example.com",
			Subject: "Subject",
			Body:    "Body",
		}
		jsonPayload, _ := json.Marshal(payload)
		task := asynq.NewTask(tasks.TypeSendEmail, jsonPayload)

		err := handler.ProcessTaskSendEmail(context.Background(), task)
		assert.NoError(t, err)
	})

	t.Run("Unmarshal Error", func(t *testing.T) {
		task := asynq.NewTask(tasks.TypeSendEmail, []byte("invalid json"))

		err := handler.ProcessTaskSendEmail(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal task payload")
	})
}
