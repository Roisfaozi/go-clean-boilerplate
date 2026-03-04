package handlers

import (
	"context"
	"fmt"

	auditEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type OutboxTaskHandler struct {
	repo usecase.AuditRepository
	log  *logrus.Logger
}

func NewOutboxTaskHandler(repo usecase.AuditRepository, log *logrus.Logger) *OutboxTaskHandler {
	return &OutboxTaskHandler{
		repo: repo,
		log:  log,
	}
}

// ProcessAuditOutbox is a background task that periodically flushes the outbox to the main audit log table
func (h *OutboxTaskHandler) ProcessAuditOutbox(ctx context.Context, t *asynq.Task) error {
	h.log.Info("Starting audit outbox synchronization...")

	// Batch size could be configurable
	batchSize := 50
	entries, err := h.repo.FindPendingOutbox(ctx, batchSize)
	if err != nil {
		h.log.WithError(err).Error("Failed to fetch pending outbox entries")
		return err
	}

	if len(entries) == 0 {
		h.log.Debug("No pending audit outbox entries found.")
		return nil
	}

	successCount := 0
	for _, entry := range entries {
		if err := h.processEntry(ctx, entry); err != nil {
			h.log.WithError(err).Errorf("Failed to process outbox entry %s", entry.ID)
			_ = h.repo.UpdateOutboxStatus(ctx, entry.ID, auditEntity.OutboxStatusFailed, err.Error())
			continue
		}
		successCount++
	}

	h.log.Infof("Audit outbox synchronization complete. Processed: %d, Success: %d", len(entries), successCount)
	return nil
}

func (h *OutboxTaskHandler) processEntry(ctx context.Context, entry *auditEntity.AuditOutbox) error {
	// Map Outbox to AuditLog
	auditLog := &auditEntity.AuditLog{
		OrganizationID: entry.OrganizationID,
		UserID:         entry.UserID,
		Action:         entry.Action,
		Entity:         entry.Entity,
		EntityID:       entry.EntityID,
		OldValues:      entry.OldValues,
		NewValues:      entry.NewValues,
		IPAddress:      entry.IPAddress,
		UserAgent:      entry.UserAgent,
		CreatedAt:      entry.CreatedAt,
	}

	// 1. Create in main audit log table
	if err := h.repo.Create(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to move outbox to main log: %w", err)
	}

	// 2. Remove from outbox (Strongly consistent move)
	if err := h.repo.DeleteOutbox(ctx, entry.ID); err != nil {
		h.log.WithError(err).Warnf("Successfully logged audit but failed to delete outbox entry %s", entry.ID)
		// We don't fail here to avoid duplicate processing if possible,
		// but the deletion failure will trigger a retry usually.
	}

	return nil
}
