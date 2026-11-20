package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	TypeSendEmail = "email:send"
)

type SendEmailPayload struct {
	To             string `json:"to"`
	Subject        string `json:"subject"`
	Body           string `json:"body"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

func NewSendEmailTask(to, subject, body string) (*asynq.Task, error) {
	return NewSendEmailTaskWithPayload(&SendEmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	})
}

func NewSendEmailTaskWithPayload(payload *SendEmailPayload) (*asynq.Task, error) {
	if payload == nil {
		return nil, fmt.Errorf("email payload is nil")
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email payload: %w", err)
	}
	return asynq.NewTask(TypeSendEmail, jsonPayload), nil
}
