package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	auditModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type AuditTaskHandler struct {
	logger  *logrus.Logger
	auditUC auditUseCase.AuditUseCase
}

func NewAuditTaskHandler(logger *logrus.Logger, auditUC auditUseCase.AuditUseCase) *AuditTaskHandler {
	return &AuditTaskHandler{
		logger:  logger,
		auditUC: auditUC,
	}
}

func (h *AuditTaskHandler) ProcessTaskAuditLog(ctx context.Context, t *asynq.Task) error {
	var payload auditModel.CreateAuditLogRequest
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal audit log payload: %w", err)
	}

	if err := h.auditUC.LogActivity(ctx, payload); err != nil {
		return fmt.Errorf("failed to log audit activity: %w", err)
	}

	return nil
}
