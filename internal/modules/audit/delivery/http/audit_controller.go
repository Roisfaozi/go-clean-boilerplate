package http

import (
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuditController struct {
	UseCase usecase.AuditUseCase
	Log     *logrus.Logger
}

func NewAuditController(uc usecase.AuditUseCase, log *logrus.Logger) *AuditController {
	return &AuditController{
		UseCase: uc,
		Log:     log,
	}
}

func (h *AuditController) GetLogsDynamic(c *gin.Context) {
	var filter querybuilder.DynamicFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.BadRequest(c, err, "Invalid filter format")
		return
	}

	logs, total, err := h.UseCase.GetLogsDynamic(c.Request.Context(), &filter)
	if err != nil {
		response.InternalServerError(c, err, "Failed to fetch logs")
		return
	}

	response.SuccessResponseWithPaging(c, logs, &response.PageMetadata{
		Page:  filter.Page,
		Limit: filter.PageSize,
		Total: total,
	})
}

func (h *AuditController) Export(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)

	// Write header
	header := []string{"ID", "UserID", "Action", "Entity", "EntityID", "OldValues", "NewValues", "IPAddress", "UserAgent", "CreatedAt"}
	if err := writer.Write(header); err != nil {
		h.Log.WithError(err).Error("Failed to write CSV header")
		response.InternalServerError(c, err, "Failed to generate CSV")
		return
	}
	writer.Flush()

	err := h.UseCase.ExportLogs(c.Request.Context(), fromDate, toDate, func(logs []model.AuditLogResponse) error {
		for _, log := range logs {
			oldVal, oldErr := json.Marshal(log.OldValues)
			if oldErr != nil {
				h.Log.WithError(oldErr).Warnf("Failed to marshal OldValues for audit log %s", log.ID)
				oldVal = []byte("null")
			}
			newVal, newErr := json.Marshal(log.NewValues)
			if newErr != nil {
				h.Log.WithError(newErr).Warnf("Failed to marshal NewValues for audit log %s", log.ID)
				newVal = []byte("null")
			}
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
		return writer.Error()
	})

	if err != nil {
		h.Log.WithError(err).Error("Failed to export logs")
		return
	}
}
