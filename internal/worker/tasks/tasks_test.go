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
