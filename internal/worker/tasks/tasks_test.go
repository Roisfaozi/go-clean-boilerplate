package tasks_test

import (
	"encoding/json"
	"math"
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

func TestNewWebhookTriggerTask(t *testing.T) {
	inputPayload := tasks.WebhookTriggerPayload{
		WebhookID: "123",
		URL:       "http://example.com",
		Secret:    "secret",
		EventType: "event",
		Payload:   `{"key":"value"}`,
	}
	task, err := tasks.NewWebhookTriggerTask(inputPayload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeWebhookTrigger, task.Type())

	var payload tasks.WebhookTriggerPayload
	err = json.Unmarshal(task.Payload(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, inputPayload.WebhookID, payload.WebhookID)
	assert.Equal(t, inputPayload.URL, payload.URL)
	assert.Equal(t, inputPayload.Secret, payload.Secret)
	assert.Equal(t, inputPayload.EventType, payload.EventType)
	assert.Equal(t, inputPayload.Payload, payload.Payload)
}

func TestNewAuditOutboxSyncTask(t *testing.T) {
	task := tasks.NewAuditOutboxSyncTask()
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeAuditOutboxSync, task.Type())
}

func TestNewAuditLogCreateTask(t *testing.T) {
	inputPayload := auditModel.CreateAuditLogRequest{
		OrganizationID: "123",
		UserID:         "user1",
		Action:         "CREATE",
		Entity:         "User",
		EntityID:       "user2",
	}
	task, err := tasks.NewAuditLogCreateTask(inputPayload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeAuditLogCreate, task.Type())

	var payload auditModel.CreateAuditLogRequest
	err = json.Unmarshal(task.Payload(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, inputPayload.OrganizationID, payload.OrganizationID)
	assert.Equal(t, inputPayload.UserID, payload.UserID)
	assert.Equal(t, inputPayload.Action, payload.Action)
	assert.Equal(t, inputPayload.Entity, payload.Entity)
	assert.Equal(t, inputPayload.EntityID, payload.EntityID)
}

func TestNewAuditLogCreateTask_JSONError(t *testing.T) {
	// Trigger JSON marshal error by using math.NaN() in OldValues
	_, err := tasks.NewAuditLogCreateTask(auditModel.CreateAuditLogRequest{
		OldValues: math.NaN(),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal audit log payload")
}

func TestNewAuditLogExportTask(t *testing.T) {
	inputPayload := auditModel.AuditLogExportPayload{
		UserID:         "user1",
		OrganizationID: "org1",
		FromDate:       "2023-01-01",
		ToDate:         "2023-12-31",
		Format:         "csv",
	}
	task, err := tasks.NewAuditLogExportTask(inputPayload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks.TypeAuditLogExport, task.Type())

	var payload auditModel.AuditLogExportPayload
	err = json.Unmarshal(task.Payload(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, inputPayload.UserID, payload.UserID)
	assert.Equal(t, inputPayload.OrganizationID, payload.OrganizationID)
	assert.Equal(t, inputPayload.FromDate, payload.FromDate)
	assert.Equal(t, inputPayload.ToDate, payload.ToDate)
	assert.Equal(t, inputPayload.Format, payload.Format)
}
