package http_test

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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation" // Import validation pkg
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

func setupRoleTestRouter(uc usecase.RoleUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	handler := roleHttp.NewRoleController(uc, logrus.New(), v)
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/roles", handler.Create)
		apiV1.GET("/roles", handler.GetAll)
		apiV1.PUT("/roles/:id", handler.Update)
		apiV1.DELETE("/roles/:id", handler.Delete)
		apiV1.POST("/roles/search", handler.GetRolesDynamic)

		apiV1.POST("/organizations/:id/roles", handler.CreateOrganizationRole)
		apiV1.GET("/organizations/:id/roles", handler.GetOrganizationRoles)
		apiV1.PUT("/organizations/:id/roles/:roleId", handler.UpdateOrganizationRole)
		apiV1.DELETE("/organizations/:id/roles/:roleId", handler.DeleteOrganizationRole)
	}
	return router
}

func TestRoleHandler_Create_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

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
	router := setupRoleTestRouter(mockUseCase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRoleHandler_Create_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	createRequest := model.CreateRoleRequest{Name: "", Description: "Administrator role"} // Invalid name
	requestBody, _ := json.Marshal(createRequest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "validation error")
	mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRoleHandler_Create_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

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
	router := setupRoleTestRouter(mockUseCase)

	expectedRoles := []model.RoleResponse{
		{ID: "1", Name: "admin"},
		{ID: "2", Name: "user"},
	}
	mockUseCase.On("GetAll", mock.Anything).Return(expectedRoles, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody response.WebResponseSuccess[[]model.RoleResponse]
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Len(t, responseBody.Data, 2)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetAll_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	mockUseCase.On("GetAll", mock.Anything).Return(nil, errors.New("some database error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Delete_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

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
	router := setupRoleTestRouter(mockUseCase)

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
	router := setupRoleTestRouter(mockUseCase)

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
	router := setupRoleTestRouter(mockUseCase)

	dynamicFilter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"Name": {Type: "contains", From: "test"},
		},
	}
	requestBody, _ := json.Marshal(dynamicFilter)

	expectedRoles := []model.RoleResponse{
		{ID: "1", Name: "test_role"},
	}
	mockUseCase.On("GetAllRolesDynamic", mock.Anything, dynamicFilter).Return(expectedRoles, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles/search", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody response.WebResponseSuccess[[]model.RoleResponse]
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Len(t, responseBody.Data, 1)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Create_XSS_Name(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	createRequest := model.CreateRoleRequest{Name: "<script>alert(1)</script>", Description: "XSS role"}
	requestBody, _ := json.Marshal(createRequest)

	// Expect sanitized input
	sanitizedRequest := model.CreateRoleRequest{Name: "&lt;script&gt;alert(1)&lt;/script&gt;", Description: "XSS role"}
	mockUseCase.On("Create", mock.Anything, &sanitizedRequest).Return(&model.RoleResponse{ID: "uuid", Name: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should return 201 Created due to sanitization
	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Update_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	roleID := "test-uuid"
	updateRequest := model.UpdateRoleRequest{Description: "Updated description"}
	requestBody, _ := json.Marshal(updateRequest)

	mockUseCase.On("Update", mock.Anything, roleID, &updateRequest).Return(&model.RoleResponse{ID: roleID, Name: "admin", Description: "Updated description"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+roleID, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Update_BindingError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	roleID := "test-uuid"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+roleID, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request body")
	mockUseCase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_Update_XSS_Sanitization(t *testing.T) {

	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	roleID := "test-uuid"
	updateRequest := model.UpdateRoleRequest{Description: "<script>alert(1)</script>"}
	requestBody, _ := json.Marshal(updateRequest)

	sanitizedRequest := model.UpdateRoleRequest{Description: "&lt;script&gt;alert(1)&lt;/script&gt;"}
	mockUseCase.On("Update", mock.Anything, roleID, &sanitizedRequest).Return(&model.RoleResponse{ID: roleID, Name: "admin", Description: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+roleID, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Since XSS is sanitized, validation passes and update occurs.
	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Update_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	roleID := "test-uuid"
	updateRequest := model.UpdateRoleRequest{Description: "Updated description"}
	requestBody, _ := json.Marshal(updateRequest)

	mockUseCase.On("Update", mock.Anything, roleID, &updateRequest).Return(nil, exception.ErrNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+roleID, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetRolesDynamic_BindingError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles/search", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "GetAllRolesDynamic", mock.Anything, mock.Anything)
}

func TestRoleHandler_GetRolesDynamic_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	dynamicFilter := &querybuilder.DynamicFilter{}
	requestBody, _ := json.Marshal(dynamicFilter)

	mockUseCase.On("GetAllRolesDynamic", mock.Anything, dynamicFilter).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles/search", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_HandleError_Variants(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	roleID := "test-uuid"

	tests := []struct {
		err          error
		expectedCode int
		expectedBodyContains string
	}{
		{exception.ErrBadRequest, http.StatusBadRequest,"failed to delete role"},
		{exception.ErrUnauthorized, http.StatusUnauthorized,"failed to delete role"},
		{exception.ErrForbidden, http.StatusForbidden,"failed to delete role"},
		{exception.ErrNotFound, http.StatusNotFound,"failed to delete role"},
		{exception.ErrConflict, http.StatusConflict,"failed to delete role"},
		{errors.New("unknown error"), http.StatusInternalServerError, "something went wrong"},
	}

	for _, tt := range tests {
		mockUseCase.ExpectedCalls = nil // Clear expected calls
		mockUseCase.On("Delete", mock.Anything, roleID).Return(tt.err).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+roleID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tt.expectedCode, w.Code, "Expected code %d for error %v", tt.expectedCode, tt.err)
		assert.Contains(t, w.Body.String(), tt.expectedBodyContains)  
	}
}

func TestRoleHandler_CreateOrganizationRole_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	createRequest := model.CreateRoleRequest{Name: "custom_role", Description: "Custom role"}
	requestBody, _ := json.Marshal(createRequest)

	mockUseCase.On("CreateForOrganization", mock.Anything, orgID, &createRequest).Return(&model.RoleResponse{ID: "role-1", Name: "custom_role"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_CreateOrganizationRole_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/roles", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateForOrganization", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_CreateOrganizationRole_ValidationFail(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	createRequest := model.CreateRoleRequest{Name: "", Description: "No name"} // invalid
	requestBody, _ := json.Marshal(createRequest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateForOrganization", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_CreateOrganizationRole_MissingOrgID(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)

	createRequest := model.CreateRoleRequest{Name: "custom_role"}
	requestBody, _ := json.Marshal(createRequest)

	// In the test router setup, if we hit a path without id, it might hit a different route.
	// But let's simulate the router matching with an empty orgID or direct invocation if needed.
	// We can test this by making a request to an endpoint where ID is somehow empty.
	// Gin's router usually doesn't allow empty param in path `/organizations//roles`, so it returns 404.
	// Alternatively, we can mock `GetOrganizationIDFromContext`.
	// The implementation checks: `orgID := c.Param("id")`. If it's an empty string, it fails.
	// Let's just create a custom route to test this specific logic.

	gin.SetMode(gin.TestMode)
	testRouter := gin.New()
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	handler := roleHttp.NewRoleController(mockUseCase, logrus.New(), v)

	// Create a route that doesn't provide the 'id' param
	testRouter.POST("/organizations/empty-id/roles", func(c *gin.Context) {
		// Override id param to empty string
		c.Params = []gin.Param{{Key: "id", Value: ""}}
		handler.CreateOrganizationRole(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/organizations/empty-id/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "organization context required")
	mockUseCase.AssertNotCalled(t, "CreateForOrganization", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_CreateOrganizationRole_XSS(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	createRequest := model.CreateRoleRequest{Name: "<script>alert(1)</script>"}
	requestBody, _ := json.Marshal(createRequest)

	sanitizedRequest := model.CreateRoleRequest{Name: "&lt;script&gt;alert(1)&lt;/script&gt;"}
	mockUseCase.On("CreateForOrganization", mock.Anything, orgID, &sanitizedRequest).Return(&model.RoleResponse{ID: "role-1", Name: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_CreateOrganizationRole_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	createRequest := model.CreateRoleRequest{Name: "custom_role"}
	requestBody, _ := json.Marshal(createRequest)

	mockUseCase.On("CreateForOrganization", mock.Anything, orgID, &createRequest).Return(nil, exception.ErrConflict)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/"+orgID+"/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetOrganizationRoles_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	expectedRoles := []model.RoleResponse{
		{ID: "role-1", Name: "custom_role_1"},
		{ID: "role-2", Name: "custom_role_2"},
	}
	mockUseCase.On("GetOrganizationRoles", mock.Anything, orgID).Return(expectedRoles, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID+"/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseBody response.WebResponseSuccess[[]model.RoleResponse]
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Len(t, responseBody.Data, 2)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_GetOrganizationRoles_MissingOrgID(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)

	gin.SetMode(gin.TestMode)
	testRouter := gin.New()
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	handler := roleHttp.NewRoleController(mockUseCase, logrus.New(), v)

	testRouter.GET("/organizations/empty-id/roles", func(c *gin.Context) {
		c.Params = []gin.Param{{Key: "id", Value: ""}}
		handler.GetOrganizationRoles(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/organizations/empty-id/roles", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "organization context required")
	mockUseCase.AssertNotCalled(t, "GetOrganizationRoles", mock.Anything, mock.Anything)
}

func TestRoleHandler_GetOrganizationRoles_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	mockUseCase.On("GetOrganizationRoles", mock.Anything, orgID).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID+"/roles", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_UpdateOrganizationRole_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	roleID := "role-1"
	updateRequest := model.UpdateRoleRequest{Description: "Updated desc"}
	requestBody, _ := json.Marshal(updateRequest)

	mockUseCase.On("UpdateForOrganization", mock.Anything, orgID, roleID, &updateRequest).Return(&model.RoleResponse{ID: roleID, Name: "custom_role", Description: "Updated desc"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID+"/roles/"+roleID, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_UpdateOrganizationRole_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	roleID := "role-1"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID+"/roles/"+roleID, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "UpdateForOrganization", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateOrganizationRole_MissingIDs(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)

	gin.SetMode(gin.TestMode)
	testRouter := gin.New()
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	handler := roleHttp.NewRoleController(mockUseCase, logrus.New(), v)

	testRouter.PUT("/organizations/empty/roles/empty", func(c *gin.Context) {
		// Test missing orgID
		c.Params = []gin.Param{{Key: "id", Value: ""}, {Key: "roleId", Value: "role-1"}}
		handler.UpdateOrganizationRole(c)
	})

	updateRequest := model.UpdateRoleRequest{Description: "Updated desc"}
	requestBody, _ := json.Marshal(updateRequest)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/organizations/empty/roles/empty", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "organization and role ID required")
	mockUseCase.AssertNotCalled(t, "UpdateForOrganization", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_UpdateOrganizationRole_XSS(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	roleID := "role-1"
	updateRequest := model.UpdateRoleRequest{Description: "<script>alert(1)</script>"}
	requestBody, _ := json.Marshal(updateRequest)

	sanitizedRequest := model.UpdateRoleRequest{Description: "&lt;script&gt;alert(1)&lt;/script&gt;"}
	mockUseCase.On("UpdateForOrganization", mock.Anything, orgID, roleID, &sanitizedRequest).Return(&model.RoleResponse{ID: roleID, Name: "custom_role", Description: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID+"/roles/"+roleID, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_UpdateOrganizationRole_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	roleID := "role-1"
	updateRequest := model.UpdateRoleRequest{Description: "Updated desc"}
	requestBody, _ := json.Marshal(updateRequest)

	mockUseCase.On("UpdateForOrganization", mock.Anything, orgID, roleID, &updateRequest).Return(nil, exception.ErrNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/organizations/"+orgID+"/roles/"+roleID, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_DeleteOrganizationRole_Success(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	roleID := "role-1"

	mockUseCase.On("DeleteForOrganization", mock.Anything, orgID, roleID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID+"/roles/"+roleID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_DeleteOrganizationRole_MissingIDs(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)

	gin.SetMode(gin.TestMode)
	testRouter := gin.New()
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	handler := roleHttp.NewRoleController(mockUseCase, logrus.New(), v)

	testRouter.DELETE("/organizations/empty/roles/empty", func(c *gin.Context) {
		// Test missing orgID
		c.Params = []gin.Param{{Key: "id", Value: ""}, {Key: "roleId", Value: "role-1"}}
		handler.DeleteOrganizationRole(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/organizations/empty/roles/empty", nil)
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "organization and role ID required")
	mockUseCase.AssertNotCalled(t, "DeleteForOrganization", mock.Anything, mock.Anything, mock.Anything)
}

func TestRoleHandler_DeleteOrganizationRole_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	orgID := "org-123"
	roleID := "superadmin-uuid" // assume we can't delete this
	mockUseCase.On("DeleteForOrganization", mock.Anything, orgID, roleID).Return(exception.ErrForbidden)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/organizations/"+orgID+"/roles/"+roleID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockUseCase.AssertExpectations(t)
}
