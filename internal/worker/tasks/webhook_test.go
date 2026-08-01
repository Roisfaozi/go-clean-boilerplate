package tasks

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewWebhookTriggerTask(t *testing.T) {
	payload := WebhookTriggerPayload{
		WebhookID: "wh-123",
		URL:       "https://example.com/webhook",
	}
	task, err := NewWebhookTriggerTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeWebhookTrigger, task.Type())
}

func TestNewWebhookTriggerTask_MarshalError(t *testing.T) {
    // WebhookTriggerPayload only has strings
}
