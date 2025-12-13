package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// NoOpWriter is a logrus.Hook that discards all log entries.
type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

// setupRouter sets up a Gin router with the RoleHandler
func setupRouter(uc usecase.RoleUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := roleHttp.NewRoleHandler(uc, logrus.New(), validator.New()) // Inject validator
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/roles", handler.Create)
		apiV1.GET("/roles", handler.GetAll)
		apiV1.DELETE("/roles/:id", handler.Delete)
		apiV1.POST("/roles/search", handler.GetRolesDynamic) // Correct handler method name
	}
	return router
}

func TestRoleHandler_Create_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	createRequest := model.CreateRoleRequest{Name: "admin", Description: "Administrator role"}
	requestBody, _ := json.Marshal(createRequest)

	mockUseCase.On("Create", mock.Anything, &createRequest).Return(&model.RoleResponse{ID: "uuid", Name: "admin"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Create_BindingError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "failed to bind request body for create role")
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRoleHandler_Create_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	createRequest := model.CreateRoleRequest{Name: "", Description: "Administrator role"} // Invalid name
	requestBody, _ := json.Marshal(createRequest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "Validation Error")
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRoleHandler_Create_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	createRequest := model.CreateRoleRequest{Name: "existing", Description: "Existing role"}
	requestBody, _ := json.Marshal(createRequest)

	mockUseCase.On("Create", mock.Anything, &createRequest).Return(nil, exception.ErrConflict)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetAll_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	expectedRoles := []*model.RoleResponse{
		{ID: "1", Name: "admin"},
		{ID: "2", Name: "user"},
	}
	mockUseCase.On("GetAll", mock.Anything).Return(expectedRoles, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Correct response struct
	var responseBody response.WebResponseSuccess[[]*model.RoleResponse]
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	// WebResponseSuccess struct has no Success bool field, existence of Data implies success or we check HTTP code
	assert.Len(t, responseBody.Data, 2)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetAll_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	mockUseCase.On("GetAll", mock.Anything).Return(nil, errors.New("some database error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Delete_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	roleID := "test-uuid"
	mockUseCase.On("Delete", mock.Anything, roleID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+roleID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Delete_NotFound(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	roleID := "non-existent-uuid"
	mockUseCase.On("Delete", mock.Anything, roleID).Return(exception.ErrNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+roleID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Delete_Forbidden(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	roleID := "superadmin-uuid"
	mockUseCase.On("Delete", mock.Anything, roleID).Return(exception.ErrForbidden)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+roleID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetAllRolesDynamic_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRouter(mockUseCase)

	dynamicFilter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"Name": {Type: "contains", From: "test"},
		},
	}
	requestBody, _ := json.Marshal(dynamicFilter)

	expectedRoles := []*model.RoleResponse{
		{ID: "1", Name: "test_role"},
	}
	mockUseCase.On("GetAllRolesDynamic", mock.Anything, dynamicFilter).Return(expectedRoles, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles/search", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	// Correct response struct
	var responseBody response.WebResponseSuccess[[]*model.RoleResponse]
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Len(t, responseBody.Data, 1)
	mockUseCase.AssertExpectations(t)
}
