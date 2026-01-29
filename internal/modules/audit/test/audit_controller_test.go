package test_test

import (
	"bytes"
	"encoding/json"
	"errors"
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
