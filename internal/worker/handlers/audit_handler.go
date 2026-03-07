package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

func (h *AuditTaskHandler) ProcessTaskAuditLogExport(ctx context.Context, t *asynq.Task) error {
	var payload auditModel.AuditLogExportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal audit log export payload: %w", err)
	}

	h.logger.Infof("Processing audit log export for user %s, org %s", payload.UserID, payload.OrganizationID)

	// Create a temporary file or a dedicated export directory
	exportDir := "exports"
	if err := os.MkdirAll(exportDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	fileName := fmt.Sprintf("audit_logs_export_%s_%d.csv", payload.UserID, time.Now().Unix())

	// For now, let's just use a fixed path to demonstrate
	filePath := filepath.Join(exportDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.WithError(err).Error("failed to close export file")
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"ID", "UserID", "Action", "Entity", "EntityID", "OldValues", "NewValues", "IPAddress", "UserAgent", "CreatedAt"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	err = h.auditUC.ExportLogs(ctx, payload.FromDate, payload.ToDate, func(logs []auditModel.AuditLogResponse) error {
		for _, log := range logs {
			oldVal, _ := json.Marshal(log.OldValues)
			newVal, _ := json.Marshal(log.NewValues)
			record := []string{
				log.ID,
				log.UserID,
				log.Action,
				log.Entity,
				log.EntityID,
				string(oldVal),
				string(newVal),
				log.IPAddress,
				log.UserAgent,
				fmt.Sprintf("%d", log.CreatedAt),
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
		writer.Flush()
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to export logs: %w", err)
	}

	h.logger.Infof("Audit log export completed: %s", filePath)
	// In a real scenario, we would upload to S3 here and notify the user via WebSocket/Email.

	return nil
}
