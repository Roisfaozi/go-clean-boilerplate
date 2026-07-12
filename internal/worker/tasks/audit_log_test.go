package tasks

import (
	"encoding/json"
	"testing"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/stretchr/testify/assert"
)

func TestNewAuditOutboxSyncTask(t *testing.T) {
	t.Run("Positive - Create AuditOutboxSync task successfully", func(t *testing.T) {
		task := NewAuditOutboxSyncTask()
		assert.NotNil(t, task)
		assert.Equal(t, TypeAuditOutboxSync, task.Type())
		assert.Nil(t, task.Payload())
	})

	// Negative and Vulnerability scenarios don't apply as this function takes no arguments and has no failure conditions.
}

func TestNewAuditLogCreateTask(t *testing.T) {
	t.Run("Positive - Create AuditLogCreate task successfully", func(t *testing.T) {
		payload := auditModel.CreateAuditLogRequest{
			Action: "test_action",
		}

		task, err := NewAuditLogCreateTask(payload)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, TypeAuditLogCreate, task.Type())

		var decodedPayload auditModel.CreateAuditLogRequest
		err = json.Unmarshal(task.Payload(), &decodedPayload)
		assert.NoError(t, err)
		assert.Equal(t, payload.Action, decodedPayload.Action)
	})

	// Since CreateAuditLogRequest struct fields are serializable, testing marshaling errors via struct is difficult.
	// Bypassing Negative & Vulnerability scenario for standard json.Marshal on struct.
}

func TestNewAuditLogExportTask(t *testing.T) {
	t.Run("Positive - Create AuditLogExport task successfully", func(t *testing.T) {
		payload := auditModel.AuditLogExportPayload{
			UserID: "user_123",
		}

		task, err := NewAuditLogExportTask(payload)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, TypeAuditLogExport, task.Type())

		var decodedPayload auditModel.AuditLogExportPayload
		err = json.Unmarshal(task.Payload(), &decodedPayload)
		assert.NoError(t, err)
		assert.Equal(t, payload.UserID, decodedPayload.UserID)
	})

	// Similar to NewAuditLogCreateTask, JSON marshaling on a standard struct is unlikely to fail.
}
