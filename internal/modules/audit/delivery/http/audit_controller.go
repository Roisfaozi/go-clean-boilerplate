package http

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AuditController struct {
	UseCase  usecase.AuditUseCase
	Validate *validator.Validate
	Log      *logrus.Logger
}

func NewAuditController(uc usecase.AuditUseCase, validate *validator.Validate, log *logrus.Logger) *AuditController {
	return &AuditController{
		UseCase:  uc,
		Validate: validate,
		Log:      log,
	}
}

// GetLogsDynamic godoc
// @Summary      Search audit logs
// @Description  Retrieves audit logs with dynamic filtering and pagination.
// @Tags         audit-logs
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        filter body querybuilder.DynamicFilter true "Dynamic filter and sort criteria"
// @Success      200  {object}  response.SwaggerAuditLogListResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid filter format"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /audit-logs/search [post]
func (h *AuditController) GetLogsDynamic(c *gin.Context) {
	var filter querybuilder.DynamicFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.BadRequest(c, err, "Invalid filter format")
		return
	}

	if err := h.Validate.Struct(filter); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
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

// Export godoc
// @Summary      Export audit logs
// @Description  Exports audit logs to CSV format within a date range.
// @Tags         audit-logs
// @Security     BearerAuth
// @Produce      text/csv
// @Param        from_date query string false "Start date (YYYY-MM-DD)"
// @Param        to_date query string false "End date (YYYY-MM-DD)"
// @Success      200  {file}  file "CSV file download"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /audit-logs/export [get]
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
				sanitizeCSVField(log.ID),
				sanitizeCSVField(log.UserID),
				sanitizeCSVField(log.Action),
				sanitizeCSVField(log.Entity),
				sanitizeCSVField(log.EntityID),
				sanitizeCSVField(string(oldVal)),
				sanitizeCSVField(string(newVal)),
				sanitizeCSVField(log.IPAddress),
				sanitizeCSVField(log.UserAgent),
				sanitizeCSVField(fmt.Sprintf("%d", log.CreatedAt)),
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

// sanitizeCSVField escapes fields to prevent CSV injection (formula injection).
// Prepend a single quote sticking to OWASP recommendation if the field starts with =, +, -, or @.
func sanitizeCSVField(val string) string {
	if len(val) > 0 {
		first := val[0]
		if first == '=' || first == '+' || first == '-' || first == '@' {
			return "'" + val
		}
	}
	return val
}
