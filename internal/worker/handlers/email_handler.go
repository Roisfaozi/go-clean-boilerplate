package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// SMTPConfig is defined locally/interface to avoid import cycle
type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromSender string
	FromEmail  string
}

type EmailTaskHandler struct {
	logger *logrus.Logger
	cfg    SMTPConfig
}

func NewEmailTaskHandler(logger *logrus.Logger, cfg SMTPConfig) *EmailTaskHandler {
	return &EmailTaskHandler{
		logger: logger,
		cfg:    cfg,
	}
}

func (h *EmailTaskHandler) ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	h.logger.WithContext(ctx).Infof("Sending real email to %s via %s:%d", payload.To, h.cfg.Host, h.cfg.Port)

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", h.cfg.FromSender, h.cfg.FromEmail))
	m.SetHeader("To", payload.To)
	m.SetHeader("Subject", payload.Subject)
	m.SetBody("text/html", payload.Body)

	d := gomail.NewDialer(h.cfg.Host, h.cfg.Port, h.cfg.Username, h.cfg.Password)

	if err := d.DialAndSend(m); err != nil {
		h.logger.WithContext(ctx).Errorf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	h.logger.Infof("SUCCESS: Email sent to %s", payload.To)
	return nil
}
