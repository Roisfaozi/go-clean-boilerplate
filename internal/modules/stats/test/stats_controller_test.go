package test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	statsHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupStatsControllerTest() (*mocks.MockStatsUseCase, *statsHttp.StatsController, *gin.Engine) {
	mockUseCase := new(mocks.MockStatsUseCase)
	controller := statsHttp.NewStatsController(mockUseCase)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Stats endpoints usually require authentication/authorization, but controller itself doesn't check it unless middleware is applied.
	// We simulate basic request context.
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-1")
		c.Next()
	})

	return mockUseCase, controller, r
}

func TestStatsController_GetSummary_Success(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/summary", controller.GetSummary)

	expectedSummary := &model.DashboardSummary{
		TotalUsers:      100,
		TotalRoles:      5,
		TotalAuditLogs:  500,
		TotalOrgMembers: 120,
	}

	mockUseCase.On("GetDashboardSummary", mock.Anything).Return(expectedSummary, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/summary", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsController_GetSummary_Error(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/summary", controller.GetSummary)

	mockUseCase.On("GetDashboardSummary", mock.Anything).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/summary", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatsController_GetActivity_Success(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/activity", controller.GetActivity)

	expectedActivity := &model.DashboardActivity{
		Points: []model.ActivityPoint{
			{Date: "2023-01-01", Audits: 10, Logins: 5},
		},
	}

	// Default days is 7
	mockUseCase.On("GetDashboardActivity", mock.Anything, 7).Return(expectedActivity, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/activity", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsController_GetActivity_CustomDays(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/activity", controller.GetActivity)

	expectedActivity := &model.DashboardActivity{
		Points: []model.ActivityPoint{},
	}

	mockUseCase.On("GetDashboardActivity", mock.Anything, 30).Return(expectedActivity, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/activity?days=30", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsController_GetActivity_Error(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/activity", controller.GetActivity)

	mockUseCase.On("GetDashboardActivity", mock.Anything, 7).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/activity", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatsController_GetInsights_Success(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/insights", controller.GetInsights)

	expectedInsights := &model.SystemInsights{
		AvgLatencyMs:   20.5,
		ErrorRate:      0.01,
		Uptime:         "99.9%",
		MostActiveRole: "admin",
	}

	mockUseCase.On("GetSystemInsights", mock.Anything).Return(expectedInsights, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/insights", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStatsController_GetInsights_Error(t *testing.T) {
	mockUseCase, controller, r := setupStatsControllerTest()
	r.GET("/stats/insights", controller.GetInsights)

	mockUseCase.On("GetSystemInsights", mock.Anything).Return(nil, errors.New("service error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/stats/insights", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
