package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	permissionHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupPermissionControllerTest() (*mocks.MockIPermissionUseCase, *permissionHttp.PermissionController) {
	mockUC := new(mocks.MockIPermissionUseCase)
	logger := logrus.New()
	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)
	controller := permissionHttp.NewPermissionController(mockUC, logger, validate)
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
	}
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest("POST", "/permission/assign-role", bytes.NewBuffer(body))

	mockUC.On("AssignRoleToUser", c.Request.Context(), "u1", "role:admin").Return(nil)

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
	}
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest("POST", "/permission/revoke-role", bytes.NewBuffer(body))

	mockUC.On("RevokeRoleFromUser", c.Request.Context(), "u1", "role:admin").Return(nil)

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
	body, _ := json.Marshal(req)
	c.Request, _ = http.NewRequest("POST", "/permission/batch-check", bytes.NewBuffer(body))

	// Simulate middleware setting user_id
	c.Set("user_id", "u1")

	results := map[string]bool{"/api/v1/users:GET": true}
	mockUC.On("BatchCheckPermission", mock.Anything, "u1", req.Items).Return(results, nil)

	controller.BatchCheck(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response body
	var resp map[string]interface{} // Using generic map to avoid model import cycling if it happens
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// data.results
	data := resp["data"].(map[string]interface{})
	res := data["results"].(map[string]interface{})
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
