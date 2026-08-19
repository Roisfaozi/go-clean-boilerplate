package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	if len(name) > 2 {
		name = name[:2] + "***"
	} else {
		name = "***"
	}
	return name + "@" + parts[1]
}

func (h *EmailTaskHandler) ProcessTaskSendEmail(ctx context.Context, task *asynq.Task) error {
	var payload tasks.SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	maskedTo := MaskEmail(payload.To)
	h.logger.WithContext(ctx).Infof("Sending real email to %s via %s:%d", maskedTo, h.cfg.Host, h.cfg.Port)

	m := gomail.NewMessage()
	m.SetHeader(headerEmailFrom, fmt.Sprintf("%s <%s>", h.cfg.FromSender, h.cfg.FromEmail))
	m.SetHeader(headerEmailTo, payload.To)
	m.SetHeader(headerEmailSubject, payload.Subject)
	m.SetBody(headerValueTextHTML, payload.Body)

	d := gomail.NewDialer(h.cfg.Host, h.cfg.Port, h.cfg.Username, h.cfg.Password)

	if err := d.DialAndSend(m); err != nil {
		h.logger.WithContext(ctx).Errorf("Failed to send email to %s: %v", maskedTo, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	h.logger.WithContext(ctx).Infof("SUCCESS: Email sent to %s", maskedTo)
	return nil
}
