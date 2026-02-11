package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

type StatsController struct {
	useCase usecase.StatsUseCase
}

func NewStatsController(useCase usecase.StatsUseCase) *StatsController {
	return &StatsController{
		useCase: useCase,
	}
}

func (h *StatsController) GetSummary(c *gin.Context) {
	res, err := h.useCase.GetDashboardSummary(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get dashboard summary")
		return
	}
	response.Success(c, res)
}

func (h *StatsController) GetActivity(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)

	res, err := h.useCase.GetDashboardActivity(c.Request.Context(), days)
	if err != nil {
		response.HandleError(c, err, "failed to get activity data")
		return
	}
	response.Success(c, res)
}

func (h *StatsController) GetInsights(c *gin.Context) {
	res, err := h.useCase.GetSystemInsights(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get system insights")
		return
	}
	response.Success(c, res)
}
