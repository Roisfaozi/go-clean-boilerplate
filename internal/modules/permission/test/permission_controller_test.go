package test

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
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupPermissionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func newTestPermissionController(mockUseCase *mocks.MockIPermissionUseCase) *permHandler.PermissionController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return permHandler.NewPermissionController(mockUseCase, log, v)
}

func TestGrantPermission_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	router := setupPermissionTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
		Domain: "global",
	}
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	data, ok := responseBody["data"].(map[string]interface{})
	assert.True(t, ok, "Response should have a 'data' object")
	assert.Equal(t, "permission granted successfully", data["message"])

	mockUseCase.AssertExpectations(t)
}

func TestGrantPermission_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	router := setupPermissionTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(`{"role": "editor",`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGrantPermission_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	router := setupPermissionTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
		Domain: "global",
	}
	mockError := errors.New("use case failed")
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(mockError)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}


// --- Handler Tests ---

func setupPermissionControllerTest() (*mocks.MockIPermissionUseCase, *permHandler.PermissionController) {
	mockUC := new(mocks.MockIPermissionUseCase)
	logger := logrus.New()
	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)
	controller := permHandler.NewPermissionController(mockUC, logger, validate)
	return mockUC, controller
}

func TestPermissionController_AssignRole(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.AssignRoleRequest{
		UserID: "u1",
		Role:   "role:admin",
		Domain: "global",
	}
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest("POST", "/permission/assign-role", bytes.NewBuffer(body))

	mockUC.On("AssignRoleToUser", c.Request.Context(), "u1", "role:admin", "global").Return(nil)

	controller.AssignRole(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_RevokeRole(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.AssignRoleRequest{
		UserID: "u1",
		Role:   "role:admin",
		Domain: "global",
	}
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest("POST", "/permission/revoke-role", bytes.NewBuffer(body))

	mockUC.On("RevokeRoleFromUser", c.Request.Context(), "u1", "role:admin", "global").Return(nil)

	controller.RevokeRole(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_BatchCheck(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/api/v1/users", Action: "GET"},
		},
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)
	reqHttp, err := http.NewRequest("POST", "/permission/batch-check", bytes.NewBuffer(body))
	require.NoError(t, err)
	c.Request = reqHttp

	// Simulate middleware setting user_id
	c.Set("user_id", "u1")

	results := map[string]bool{"/api/v1/users:GET": true}
	mockUC.On("BatchCheckPermission", mock.Anything, "u1", req.Items).Return(results, nil)

	controller.BatchCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response body
	var resp map[string]interface{} // Using generic map to avoid model import cycling if it happens
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// data.results
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "data field should be a map")
	res, ok := data["results"].(map[string]interface{})
	require.True(t, ok, "results field should be a map")
	assert.Equal(t, true, res["/api/v1/users:GET"])
}

func TestPermissionController_BatchCheck_Unauthorized(t *testing.T) {
	_, controller := setupPermissionControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/api/v1/test", Action: "GET"},
		},
	}
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest("POST", "/permission/batch-check", bytes.NewBuffer(body))
	// NO user_id set

	controller.BatchCheck(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}