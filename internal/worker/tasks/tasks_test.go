package tasks_test

import (
	"encoding/json"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/stretchr/testify/assert"
)

func TestNewSendEmailTask(t *testing.T) {
	to := "test@example.com"
	subject := "Test Subject"
	body := "Test Body"

	task, err := tasks.NewSendEmailTask(to, subject, body)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeSendEmail, task.Type())

	var payload tasks.SendEmailPayload
	err = json.Unmarshal(task.Payload(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, to, payload.To)
	assert.Equal(t, subject, payload.Subject)
	assert.Equal(t, body, payload.Body)
	assert.Empty(t, payload.IdempotencyKey)
}

func TestNewSendEmailTaskWithPayload(t *testing.T) {
	t.Run("preserves idempotency key", func(t *testing.T) {
		in := &tasks.SendEmailPayload{
			To:             "test@example.com",
			Subject:        "Reset Password",
			Body:           "<p>Reset</p>",
			IdempotencyKey: "reset-user-1-abc",
		}

		task, err := tasks.NewSendEmailTaskWithPayload(in)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, tasks.TypeSendEmail, task.Type())

		var decoded tasks.SendEmailPayload
		err = json.Unmarshal(task.Payload(), &decoded)
		assert.NoError(t, err)
		assert.Equal(t, *in, decoded)
	})

	t.Run("returns error on nil payload", func(t *testing.T) {
		task, err := tasks.NewSendEmailTaskWithPayload(nil)
		assert.Error(t, err)
		assert.Nil(t, task)
	})
}

func TestNewWebhookTriggerTask_PreservesEventID(t *testing.T) {
	in := tasks.WebhookTriggerPayload{
		WebhookID: "wh-1",
		URL:       "https://example.test/hook",
		Secret:    "sec",
		EventType: "user.created",
		Payload:   `{"id":"1"}`,
		EventID:   "evt-123",
	}

	task, err := tasks.NewWebhookTriggerTask(in)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeWebhookTrigger, task.Type())

	var decoded tasks.WebhookTriggerPayload
	err = json.Unmarshal(task.Payload(), &decoded)
	assert.NoError(t, err)
	assert.Equal(t, in, decoded)
}

func TestCleanupSoftDeletedEntitiesPayload(t *testing.T) {
	payload := tasks.CleanupSoftDeletedEntitiesPayload{
		RetentionDays: 30,
	}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decoded tasks.CleanupSoftDeletedEntitiesPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.RetentionDays, decoded.RetentionDays)
}

func TestPruneAuditLogsPayload(t *testing.T) {
	payload := tasks.PruneAuditLogsPayload{
		RetentionDays: 180,
	}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var decoded tasks.PruneAuditLogsPayload
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, payload.RetentionDays, decoded.RetentionDays)
}
