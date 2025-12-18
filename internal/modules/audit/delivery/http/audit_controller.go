package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuditHandler struct {
	UseCase usecase.AuditUseCase
	Log     *logrus.Logger
}

func NewAuditHandler(uc usecase.AuditUseCase, log *logrus.Logger) *AuditHandler {
	return &AuditHandler{
		UseCase: uc,
		Log:     log,
	}
}

// GetLogsDynamic handles dynamic search for audit logs
func (h *AuditHandler) GetLogsDynamic(c *gin.Context) {
	var filter querybuilder.DynamicFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.BadRequest(c, err, "Invalid filter format")
		return
	}

	logs, err := h.UseCase.GetLogsDynamic(c.Request.Context(), &filter)
	if err != nil {
		response.InternalServerError(c, err, "Failed to fetch logs")
		return
	}

	response.Success(c, logs)
}
