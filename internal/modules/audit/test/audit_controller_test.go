package test_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupAuditTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func newTestAuditController(mockUC usecase.AuditUseCase) *auditHttp.AuditController {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return auditHttp.NewAuditController(mockUC, v, logger)
}

func TestGetLogsDynamicController(t *testing.T) {
	mockUC := new(mocks.MockAuditUseCase)
	handler := newTestAuditController(mockUC)
	router := setupAuditTestRouter()
	router.POST("/audit-logs/search", handler.GetLogsDynamic)

	t.Run("Success", func(t *testing.T) {
		filter := querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{"user_id": {Type: "equals", From: "u1"}},
		}
		body, _ := json.Marshal(filter)

		respData := []model.AuditLogResponse{
			{ID: "1", UserID: "u1"},
		}
		mockUC.On("GetLogsDynamic", mock.Anything, &filter).Return(respData, int64(1), nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/audit-logs/search", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var webResp response.WebResponseSuccess[[]model.AuditLogResponse]
		err := json.Unmarshal(w.Body.Bytes(), &webResp)
		assert.NoError(t, err, "Failed to unmarshal response")
		assert.Len(t, webResp.Data, 1)
		assert.Equal(t, int64(1), webResp.Paging.Total)
		mockUC.AssertExpectations(t)
	})

	t.Run("Bind Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/audit-logs/search", bytes.NewBufferString("{invalid json"))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		mockUC.ExpectedCalls = nil
		filter := querybuilder.DynamicFilter{}
		body, _ := json.Marshal(filter)

		mockUC.On("GetLogsDynamic", mock.Anything, &filter).Return(nil, int64(0), errors.New("fail"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/audit-logs/search", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAuditController_Export_Serialization(t *testing.T) {
	mockUC := new(mocks.MockAuditUseCase)
	controller := newTestAuditController(mockUC)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest("GET", "/audit/export?from_date=2023-01-01&to_date=2023-01-31", nil)

	fmt.Println("Setting up mock expectations...")
	mockUC.On("ExportLogs", mock.Anything, "2023-01-01", "2023-01-31", mock.Anything).
		Run(func(args mock.Arguments) {
			fmt.Println("Mock ExportLogs called!")
			iterator := args.Get(3).(func([]model.AuditLogResponse) error)
			logs := []model.AuditLogResponse{
				{
					ID:        "log-1",
					UserID:    "user-1",
					Action:    "LOGIN",
					OldValues: map[string]interface{}{"a": 1},
					NewValues: map[string]interface{}{"b": 2},
					CreatedAt: 1672531200,
				},
			}
			if err := iterator(logs); err != nil {
				fmt.Printf("Iterator returned error: %v\n", err)
			} else {
				fmt.Println("Iterator success")
			}
		}).Return(nil)

	fmt.Println("Calling controller.Export...")
	controller.Export(c)
	fmt.Println("Controller returned.")

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	if !assert.Contains(t, body, "ID,UserID") {
		t.Log("Missing CSV Header")
	}

	if !assert.Contains(t, body, "log-1") {
		t.Log("Missing CSV Record ID")
	}

	assert.Contains(t, body, "LOGIN")
	assert.Contains(t, body, `""a"":1`)
}

func TestAuditController_Export_CSVInjection(t *testing.T) {
	mockUC := new(mocks.MockAuditUseCase)
	controller := newTestAuditController(mockUC)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("GET", "/audit/export", nil)
	require.NoError(t, err)
	c.Request = req

	mockUC.On("ExportLogs", mock.Anything, "", "", mock.Anything).
		Run(func(args mock.Arguments) {
			iterator, ok := args.Get(3).(func([]model.AuditLogResponse) error)
			require.True(t, ok, "fourth argument should be an iterator function")
			logs := []model.AuditLogResponse{
				{
					ID:     "log-bad",
					UserID: "=cmd|' /C calc'!A0",
					Action: "HACK",
				},
			}
			err := iterator(logs)
			assert.NoError(t, err)
		}).Return(nil)

	controller.Export(c)

	csvOutput := w.Body.String()
	assert.Contains(t, csvOutput, "ID,UserID,Action", "CSV header missing")
	assert.Contains(t, csvOutput, "'=cmd|' /C calc'!A0", "Malicious payload should be sanitized/escaped")
	assert.NotContains(t, csvOutput, "\n=cmd|' /C calc'!A0", "Unsafe payload found (start of line)")
	assert.NotContains(t, csvOutput, ",=cmd|' /C calc'!A0", "Unsafe payload found (after comma)")
}

func TestAuditController_GetLogsDynamic_XSS(t *testing.T) {
	mockUC := new(mocks.MockAuditUseCase)
	controller := newTestAuditController(mockUC)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload := querybuilder.DynamicFilter{
		Page:     1,
		PageSize: 10,
		Sort: &[]querybuilder.SortModel{
			{
				ColId: "<script>alert(1)</script>",
				Sort:  "asc",
			},
		},
	}

	jsonBytes, _ := json.Marshal(payload)
	c.Request, _ = http.NewRequest("POST", "/audit/search", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	controller.GetLogsDynamic(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "validation failed")
}
