package test_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	permHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestGrantPermission_Success(t *testing.T) {
	mockUseCase := new(mocks.IPermissionUseCase)
	handler := permHandler.NewPermissionController(mockUseCase, validator.New(), logrus.New())
	router := setupTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
	}
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	data, ok := responseBody["data"].(map[string]interface{})
	assert.True(t, ok, "Response should have a 'data' object")
	assert.Equal(t, "Permission granted successfully", data["message"])

	mockUseCase.AssertExpectations(t)
}

func TestGrantPermission_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.IPermissionUseCase)
	handler := permHandler.NewPermissionController(mockUseCase, validator.New(), logrus.New())
	router := setupTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(`{"role": "editor",`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "GrantPermissionToRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGrantPermission_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.IPermissionUseCase)
	handler := permHandler.NewPermissionController(mockUseCase, validator.New(), logrus.New())
	router := setupTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
	}
	mockError := errors.New("use case failed")
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method).Return(mockError)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}
