package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/sirupsen/logrus"
)

type auditUseCase struct {
	repo AuditRepository
	log  *logrus.Logger
}

func NewAuditUseCase(repo AuditRepository, log *logrus.Logger) AuditUseCase {
	return &auditUseCase{
		repo: repo,
		log:  log,
	}
}

func (uc *auditUseCase) LogActivity(ctx context.Context, req model.CreateAuditLogRequest) error {
	// Validation: Ensure mandatory fields are present
	if req.UserID == "" || req.Action == "" || req.Entity == "" {
		return fmt.Errorf("missing required fields for audit log: UserID, Action, and Entity are mandatory")
	}

	oldValJSON, _ := json.Marshal(req.OldValues)
	newValJSON, _ := json.Marshal(req.NewValues)

	logEntity := &entity.AuditLog{
		UserID:    req.UserID,
		Action:    req.Action,
		Entity:    req.Entity,
		EntityID:  req.EntityID,
		OldValues: string(oldValJSON),
		NewValues: string(newValJSON),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	if err := uc.repo.Create(ctx, logEntity); err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to create audit log")
		return err
	}
	return nil
}

func (uc *auditUseCase) GetLogsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.AuditLogResponse, error) {
	logs, err := uc.repo.FindAllDynamic(ctx, filter)
	if err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to fetch audit logs")
		return nil, err
	}

	var response []model.AuditLogResponse
	for _, log := range logs {
		var oldVal, newVal interface{}
		_ = json.Unmarshal([]byte(log.OldValues), &oldVal)
		_ = json.Unmarshal([]byte(log.NewValues), &newVal)

		response = append(response, model.AuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    log.Action,
			Entity:    log.Entity,
			EntityID:  log.EntityID,
			OldValues: oldVal,
			NewValues: newVal,
			IPAddress: log.IPAddress,
			UserAgent: log.UserAgent,
			CreatedAt: log.CreatedAt,
		})
	}
	return response, nil
}
