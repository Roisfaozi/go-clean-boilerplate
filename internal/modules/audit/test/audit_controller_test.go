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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(handler *auditHttp.AuditController) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/audit-logs/search", handler.GetLogsDynamic)
	return r
}

func TestGetLogsDynamicController(t *testing.T) {
	mockUC := new(mocks.MockAuditUseCase)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := auditHttp.NewAuditController(mockUC, logger)
	router := setupRouter(handler)

	t.Run("Success", func(t *testing.T) {
		filter := querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{"user_id": {Type: "equals", From: "u1"}},
		}
		body, _ := json.Marshal(filter)

		respData := []model.AuditLogResponse{
			{ID: "1", UserID: "u1"},
		}
		mockUC.On("GetLogsDynamic", mock.Anything, &filter).Return(respData, nil)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/audit-logs/search", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var webResp response.WebResponseSuccess[[]model.AuditLogResponse]
		json.Unmarshal(w.Body.Bytes(), &webResp)
		assert.Len(t, webResp.Data, 1)
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

		mockUC.On("GetLogsDynamic", mock.Anything, &filter).Return(nil, errors.New("fail"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/audit-logs/search", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
