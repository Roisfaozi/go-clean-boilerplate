package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	authRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
)

type CleanupTaskHandler struct {
	authRepo  authRepo.TokenRepository
	userRepo  userRepo.UserRepository
	auditRepo auditUseCase.AuditRepository // Use interface defined in usecase package
	log       *logrus.Logger
}

func NewCleanupTaskHandler(
	authRepo authRepo.TokenRepository,
	userRepo userRepo.UserRepository,
	auditRepo auditUseCase.AuditRepository,
	log *logrus.Logger,
) *CleanupTaskHandler {
	return &CleanupTaskHandler{
		authRepo:  authRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		log:       log,
	}
}

// ProcessCleanupExpiredTokens deletes expired password reset tokens
func (h *CleanupTaskHandler) ProcessCleanupExpiredTokens(ctx context.Context, task *asynq.Task) error {
	h.log.Info("Starting cleanup of expired reset tokens")
	if err := h.authRepo.DeleteExpiredResetTokens(ctx); err != nil {
		telemetry.CleanupTasksTotal.WithLabelValues("expired_tokens", "failed").Inc()
		h.log.WithError(err).Error("Failed to cleanup expired reset tokens")
		return err
	}
	telemetry.CleanupTasksTotal.WithLabelValues("expired_tokens", "success").Inc()
	h.log.Info("Completed cleanup of expired reset tokens")
	return nil
}

// ProcessCleanupSoftDeletedEntities permanently deletes users soft-deleted longer than retention period
func (h *CleanupTaskHandler) ProcessCleanupSoftDeletedEntities(ctx context.Context, task *asynq.Task) error {
	var payload tasks.CleanupSoftDeletedEntitiesPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal cleanup payload: %w", err)
	}

	h.log.Infof("Starting hard delete of users soft-deleted more than %d days ago", payload.RetentionDays)
	
	if err := h.userRepo.HardDeleteSoftDeletedUsers(ctx, payload.RetentionDays); err != nil {
		telemetry.CleanupTasksTotal.WithLabelValues("soft_deleted_entities", "failed").Inc()
		h.log.WithError(err).Error("Failed to hard delete users")
		return err
	}
	
	telemetry.CleanupTasksTotal.WithLabelValues("soft_deleted_entities", "success").Inc()
	h.log.Info("Completed hard delete of old users")
	return nil
}

// ProcessPruneAuditLogs deletes audit logs older than retention period
func (h *CleanupTaskHandler) ProcessPruneAuditLogs(ctx context.Context, task *asynq.Task) error {
	var payload tasks.PruneAuditLogsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal prune logs payload: %w", err)
	}

	h.log.Infof("Starting prune of audit logs older than %d days", payload.RetentionDays)

	// Calculate cutoff timestamp (unix milli)
	cutoff := time.Now().AddDate(0, 0, -payload.RetentionDays).UnixMilli()

	if err := h.auditRepo.DeleteLogsOlderThan(ctx, cutoff); err != nil {
		telemetry.CleanupTasksTotal.WithLabelValues("prune_audit_logs", "failed").Inc()
		h.log.WithError(err).Error("Failed to prune audit logs")
		return err
	}

	telemetry.CleanupTasksTotal.WithLabelValues("prune_audit_logs", "success").Inc()
	h.log.Info("Completed prune of audit logs")
	return nil
}
