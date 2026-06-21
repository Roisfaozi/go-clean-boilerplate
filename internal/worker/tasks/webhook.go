package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	TypeWebhookTrigger = "webhook:trigger"
)

type WebhookTriggerPayload struct {
	WebhookID string `json:"webhook_id"`
	URL       string `json:"url"`
	Secret    string `json:"secret"`
	EventType string `json:"event_type"`
	Payload   string `json:"payload"` // JSON string
}

func NewWebhookTriggerTask(payload WebhookTriggerPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	return asynq.NewTask(TypeWebhookTrigger, data), nil
}
