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
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewSendEmailTask(to, subject, body string) (*asynq.Task, error) {
	payload := &SendEmailPayload{
		To:      to,
		Subject: subject,
		Body:    body,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal email payload: %w", err)
	}
	return asynq.NewTask(TypeSendEmail, jsonPayload), nil
}
