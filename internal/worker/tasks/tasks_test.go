package tasks_test

import (
	"encoding/json"
	"testing"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
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

func TestNewAuditOutboxSyncTask(t *testing.T) {
	task := tasks.NewAuditOutboxSyncTask()
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeAuditOutboxSync, task.Type())
}

func TestNewAuditLogCreateTask_Success(t *testing.T) {
	payload := auditModel.CreateAuditLogRequest{
		UserID: "user-1",
		Action: "create",
		Entity: "user",
	}
	task, err := tasks.NewAuditLogCreateTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeAuditLogCreate, task.Type())
}

func TestNewAuditLogCreateTask_MarshalError(t *testing.T) {
	payload := auditModel.CreateAuditLogRequest{
		OldValues: make(chan int), // Unsupported type for json.Marshal
	}
	task, err := tasks.NewAuditLogCreateTask(payload)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestNewAuditLogExportTask(t *testing.T) {
	payload := auditModel.AuditLogExportPayload{
		UserID: "user-1",
	}
	task, err := tasks.NewAuditLogExportTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeAuditLogExport, task.Type())
}

func TestNewWebhookTriggerTask(t *testing.T) {
	payload := tasks.WebhookTriggerPayload{
		WebhookID: "wh-1",
		URL:       "http://example.com",
	}
	task, err := tasks.NewWebhookTriggerTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeWebhookTrigger, task.Type())
}
