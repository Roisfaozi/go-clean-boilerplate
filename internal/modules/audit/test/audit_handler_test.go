package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuditControllerTest() (*mocks.MockAuditUseCase, *auditHttp.AuditController) {
	mockUC := new(mocks.MockAuditUseCase)
	logger := logrus.New()
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	controller := auditHttp.NewAuditController(mockUC, v, logger)
	return mockUC, controller
}

func TestAuditController_Export_Serialization(t *testing.T) {
	mockUC, controller := setupAuditControllerTest()

	// Setup Gin
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Mock Request
	c.Request, _ = http.NewRequest("GET", "/audit/export?from_date=2023-01-01&to_date=2023-01-31", nil)

	// Mock UseCase behavior
	fmt.Println("Setting up mock expectations...")
	mockUC.On("ExportLogs", mock.Anything, "2023-01-01", "2023-01-31", mock.Anything).
		Run(func(args mock.Arguments) {
			fmt.Println("Mock ExportLogs called!")
			// Execute the callback with sample data
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
	mockUC, controller := setupAuditControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/audit/export", nil)

	mockUC.On("ExportLogs", mock.Anything, "", "", mock.Anything).
		Run(func(args mock.Arguments) {
			iterator := args.Get(3).(func([]model.AuditLogResponse) error)
			logs := []model.AuditLogResponse{
				{
					ID:     "log-bad",
					UserID: "=cmd|' /C calc'!A0",
					Action: "HACK",
				},
			}
			_ = iterator(logs)
		}).Return(nil)

	controller.Export(c)

	csvOutput := w.Body.String()
	assert.Contains(t, csvOutput, "ID,UserID,Action", "CSV header missing")
	assert.Contains(t, csvOutput, "'=cmd|' /C calc'!A0", "Malicious payload should be sanitized/escaped")
	assert.NotContains(t, csvOutput, "\n=cmd|' /C calc'!A0", "Unsafe payload found (start of line)")
	assert.NotContains(t, csvOutput, ",=cmd|' /C calc'!A0", "Unsafe payload found (after comma)")
}

func TestAuditController_GetLogsDynamic_XSS(t *testing.T) {
	_, controller := setupAuditControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Payload with XSS in Sort ColId
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

	// The controller should call Validate.Struct(filter) and fail
	controller.GetLogsDynamic(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	// Expect validation error details
	assert.Contains(t, w.Body.String(), "validation failed")
}
