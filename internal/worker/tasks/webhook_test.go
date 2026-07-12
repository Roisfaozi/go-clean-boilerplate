package tasks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebhookTriggerTask(t *testing.T) {
	t.Run("Positive - Create WebhookTrigger task successfully", func(t *testing.T) {
		payload := WebhookTriggerPayload{
			WebhookID: "webhook_123",
			URL:       "https://example.com/webhook",
			EventType: "test_event",
		}

		task, err := NewWebhookTriggerTask(payload)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, TypeWebhookTrigger, task.Type())

		var decodedPayload WebhookTriggerPayload
		err = json.Unmarshal(task.Payload(), &decodedPayload)
		assert.NoError(t, err)
		assert.Equal(t, payload.WebhookID, decodedPayload.WebhookID)
		assert.Equal(t, payload.URL, decodedPayload.URL)
		assert.Equal(t, payload.EventType, decodedPayload.EventType)
	})

	t.Run("Edge - Create task with empty payload", func(t *testing.T) {
		payload := WebhookTriggerPayload{}

		task, err := NewWebhookTriggerTask(payload)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, TypeWebhookTrigger, task.Type())
	})

	// Similar to Audit logs, testing json.Marshal failure is not practical here as it's standard struct to JSON.
}
