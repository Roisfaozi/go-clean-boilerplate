package test_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	roleHandler "github.com/Roisfaozi/casbin-db/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestRoleHandler_Create_Success(t *testing.T) {
	mockUseCase := new(mocks.RoleUseCase)
	handler := roleHandler.NewRoleHandler(mockUseCase, logrus.New())
	router := setupTestRouter()
	router.POST("/roles", handler.Create)

	reqBody := &model.CreateRoleRequest{
		Name: "new_role",
	}
	resBody := &model.RoleResponse{
		ID:   "role-123",
		Name: "new_role",
	}

	mockUseCase.On("Create", mock.Anything, reqBody).Return(resBody, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Create_BindingError(t *testing.T) {
	mockUseCase := new(mocks.RoleUseCase)
	handler := roleHandler.NewRoleHandler(mockUseCase, logrus.New())
	router := setupTestRouter()
	router.POST("/roles", handler.Create)

	req, _ := http.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(`{"name":`)) // Invalid JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRoleHandler_Create_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.RoleUseCase)
	handler := roleHandler.NewRoleHandler(mockUseCase, logrus.New())
	router := setupTestRouter()
	router.POST("/roles", handler.Create)

	reqBody := &model.CreateRoleRequest{
		Name: "existing_role",
	}
	mockUseCase.On("Create", mock.Anything, reqBody).Return(nil, exception.ErrConflict)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetAll_Success(t *testing.T) {
	mockUseCase := new(mocks.RoleUseCase)
	handler := roleHandler.NewRoleHandler(mockUseCase, logrus.New())
	router := setupTestRouter()
	router.GET("/roles", handler.GetAll)

	expectedRoles := []model.RoleResponse{
		{ID: "role-1", Name: "admin"},
		{ID: "role-2", Name: "editor"},
	}
	mockUseCase.On("GetAll", mock.Anything).Return(expectedRoles, nil)

	req, _ := http.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &responseBody)

	data, _ := responseBody["data"].([]interface{})
	assert.Len(t, data, 2)

	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetAll_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.RoleUseCase)
	handler := roleHandler.NewRoleHandler(mockUseCase, logrus.New())
	router := setupTestRouter()
	router.GET("/roles", handler.GetAll)

	mockUseCase.On("GetAll", mock.Anything).Return(nil, errors.New("some database error"))

	req, _ := http.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}
