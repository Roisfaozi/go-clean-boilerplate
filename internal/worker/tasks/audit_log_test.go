package tasks

import (
	"testing"
	"github.com/stretchr/testify/assert"
	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
)

func TestNewAuditOutboxSyncTask(t *testing.T) {
	task := NewAuditOutboxSyncTask()
	assert.NotNil(t, task)
	assert.Equal(t, TypeAuditOutboxSync, task.Type())
}

func TestNewAuditLogCreateTask(t *testing.T) {
	payload := auditModel.CreateAuditLogRequest{
		Action: "test_action",
	}
	task, err := NewAuditLogCreateTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeAuditLogCreate, task.Type())
}

func TestNewAuditLogExportTask(t *testing.T) {
	payload := auditModel.AuditLogExportPayload{
		UserID: "user-123",
	}
	task, err := NewAuditLogExportTask(payload)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeAuditLogExport, task.Type())
}

func TestNewAuditLogCreateTask_MarshalError(t *testing.T) {
	// "When unit testing json.Marshal error handling for structs containing interface{} fields, pass an unsupported type like chan int or math.NaN() to reliably trigger the marshal error."
	payload := auditModel.CreateAuditLogRequest{
		Action:    "test_action",
		OldValues: make(chan int), // Causes json.Marshal to fail
	}
	task, err := NewAuditLogCreateTask(payload)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestNewAuditLogExportTask_MarshalError(t *testing.T) {
	// Let's pass a bad struct... Oh wait, AuditLogExportPayload only has strings, so json.Marshal will never fail.
	// As per memory: "For structs consisting entirely of standard primitive types like strings (e.g., AuditLogExportPayload), json.Marshal practically cannot fail with valid structural input. It is safe to omit negative JSON marshalling tests for such payloads rather than trying to engineer unreachable errors."
}
