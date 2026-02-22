package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/sirupsen/logrus"
)

var (
	exportBatchSize = 1000
)

type auditUseCase struct {
	repo AuditRepository
	log  *logrus.Logger
	ws   ws.Manager
}

func NewAuditUseCase(repo AuditRepository, log *logrus.Logger, ws ws.Manager) AuditUseCase {
	return &auditUseCase{
		repo: repo,
		log:  log,
		ws:   ws,
	}
}

func (uc *auditUseCase) LogActivity(ctx context.Context, req model.CreateAuditLogRequest) error {
	// Validation: Ensure mandatory fields are present
	if req.UserID == "" || req.Action == "" || req.Entity == "" {
		return fmt.Errorf("missing required fields for audit log: UserID, Action, and Entity are mandatory")
	}

	orgID := database.GetOrganizationID(ctx)
	if req.OrganizationID != "" {
		orgID = req.OrganizationID
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

	if orgID != "" {
		logEntity.OrganizationID = &orgID
	}

	if err := uc.repo.Create(ctx, logEntity); err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to create audit log")
		return err
	}

	// Broadcast event
	// Response model mapping logic here or simple map
	eventData := model.AuditLogResponse{
		ID:             logEntity.ID,
		OrganizationID: logEntity.OrganizationID,
		UserID:         logEntity.UserID,
		Action:         logEntity.Action,
		Entity:         logEntity.Entity,
		EntityID:       logEntity.EntityID,
		OldValues:      req.OldValues,
		NewValues:      req.NewValues,
		IPAddress:      logEntity.IPAddress,
		UserAgent:      logEntity.UserAgent,
		CreatedAt:      logEntity.CreatedAt,
	}

	msg, err := json.Marshal(eventData)
	if err == nil && uc.ws != nil {
		uc.ws.BroadcastToChannel("audit", msg)
	} else if err != nil {
		uc.log.WithContext(ctx).Warnf("Failed to marshal audit log event: %v", err)
	}

	return nil
}

func (uc *auditUseCase) GetLogsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.AuditLogResponse, int64, error) {
	logs, total, err := uc.repo.FindAllDynamic(ctx, filter)
	if err != nil {
		uc.log.WithContext(ctx).WithError(err).Error("Failed to fetch audit logs")
		return nil, 0, err
	}

	var response []model.AuditLogResponse
	for _, log := range logs {
		var oldVal, newVal interface{}
		_ = json.Unmarshal([]byte(log.OldValues), &oldVal)
		_ = json.Unmarshal([]byte(log.NewValues), &newVal)

		response = append(response, model.AuditLogResponse{
			ID:             log.ID,
			OrganizationID: log.OrganizationID,
			UserID:         log.UserID,
			Action:         log.Action,
			Entity:         log.Entity,
			EntityID:       log.EntityID,
			OldValues:      oldVal,
			NewValues:      newVal,
			IPAddress:      log.IPAddress,
			UserAgent:      log.UserAgent,
			CreatedAt:      log.CreatedAt,
		})
	}
	return response, total, nil
}

func (uc *auditUseCase) ExportLogs(ctx context.Context, fromDate, toDate string, process func([]model.AuditLogResponse) error) error {
	var startTime, endTime int64

	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			return fmt.Errorf("invalid from_date format, expected YYYY-MM-DD")
		}
		startTime = t.UnixMilli()
	}

	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			return fmt.Errorf("invalid to_date format, expected YYYY-MM-DD")
		}
		// End of the day
		endTime = t.Add(24 * time.Hour).UnixMilli()
	}

	batchSize := exportBatchSize

	return uc.repo.FindAllInBatches(ctx, startTime, endTime, batchSize, func(logs []*entity.AuditLog) error {
		var response []model.AuditLogResponse
		for _, log := range logs {
			var oldVal, newVal interface{}
			if err := json.Unmarshal([]byte(log.OldValues), &oldVal); err != nil {
				uc.log.WithError(err).Warnf("Failed to unmarshal OldValues for audit log %s", log.ID)
			}
			if err := json.Unmarshal([]byte(log.NewValues), &newVal); err != nil {
				uc.log.WithError(err).Warnf("Failed to unmarshal NewValues for audit log %s", log.ID)
			}
			response = append(response, model.AuditLogResponse{
				ID:             log.ID,
				OrganizationID: log.OrganizationID,
				UserID:         log.UserID,
				Action:         log.Action,
				Entity:         log.Entity,
				EntityID:       log.EntityID,
				OldValues:      oldVal,
				NewValues:      newVal,
				IPAddress:      log.IPAddress,
				UserAgent:      log.UserAgent,
				CreatedAt:      log.CreatedAt,
			})
		}
		return process(response)
	})
}
