package http

import (
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
