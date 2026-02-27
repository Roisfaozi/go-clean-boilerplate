package http

import (
	"strconv"

	_ "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
)

type StatsController struct {
	useCase usecase.StatsUseCase
}

func NewStatsController(useCase usecase.StatsUseCase) *StatsController {
	return &StatsController{
		useCase: useCase,
	}
}

// GetSummary godoc
// @Summary      Get dashboard summary
// @Description  Returns high-level system statistics (total users, roles, audit logs, etc.)
// @Tags         stats
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.DashboardSummary}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /stats/summary [get]
func (h *StatsController) GetSummary(c *gin.Context) {
	res, err := h.useCase.GetDashboardSummary(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get dashboard summary")
		return
	}
	response.Success(c, res)
}

// GetActivity godoc
// @Summary      Get activity chart data
// @Description  Returns daily activity metrics (logins, audit events) for a given period.
// @Tags         stats
// @Security     BearerAuth
// @Produce      json
// @Param        days query int false "Number of days to retrieve (default 7)"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.DashboardActivity}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /stats/activity [get]
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

// GetInsights godoc
// @Summary      Get system health insights
// @Description  Returns system-level performance metrics (latency, error rate, uptime).
// @Tags         stats
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper{data=model.SystemInsights}
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /stats/insights [get]
func (h *StatsController) GetInsights(c *gin.Context) {
	res, err := h.useCase.GetSystemInsights(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get system insights")
		return
	}
	response.Success(c, res)
}
