package tasks

import (
	"encoding/json"
	"fmt"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/hibiken/asynq"
)

const (
	TypeAuditLogCreate  = "audit_log:create"
	TypeAuditOutboxSync = "audit_log:outbox_sync"
)

func NewAuditOutboxSyncTask() *asynq.Task {
	return asynq.NewTask(TypeAuditOutboxSync, nil)
}

func NewAuditLogCreateTask(payload auditModel.CreateAuditLogRequest) (*asynq.Task, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal audit log payload: %w", err)
	}

	return asynq.NewTask(TypeAuditLogCreate, jsonPayload), nil
}
