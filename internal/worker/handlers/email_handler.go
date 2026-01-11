package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type EmailTaskHandler struct {
	logger *logrus.Logger
}

func NewEmailTaskHandler(logger *logrus.Logger) *EmailTaskHandler {
	return &EmailTaskHandler{
		logger: logger,
	}
}

func (h *EmailTaskHandler) ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	h.logger.WithContext(ctx).Infof("Processing task: sending email to %s", payload.To)

	// TODO: Integrate actual email sending logic here (e.g. SMTP, SendGrid, SES)
	h.logger.Infof("SIMULATION: Email sent to %s with subject: %s", payload.To, payload.Subject)

	return nil
}
