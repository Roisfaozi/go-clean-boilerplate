package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	permHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func wrapWithTestUser(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := delivery.SetContextValue(r.Context(), delivery.UserIDKey, "test-user-id")
		h(w, r.WithContext(ctx))
	})
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
	mux := http.NewServeMux()
	mux.Handle("POST /permissions/grant", wrapWithTestUser(handler.HTTPGrantPermission))

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
		Domain: "global",
	}
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

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
	mux := http.NewServeMux()
	mux.Handle("POST /permissions/grant", wrapWithTestUser(handler.HTTPGrantPermission))

	req := httptest.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(`{"role": "editor",`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGrantPermission_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	mux := http.NewServeMux()
	mux.Handle("POST /permissions/grant", wrapWithTestUser(handler.HTTPGrantPermission))

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
		Domain: "global",
	}
	mockError := errors.New("use case failed")
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(mockError)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAddRoleInheritance_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	mux := http.NewServeMux()
	mux.Handle("POST /permissions/inheritance", wrapWithTestUser(handler.HTTPAddRoleInheritance))

	reqBody := model.RoleInheritanceRequest{
		ChildRole:  "editor",
		ParentRole: "viewer",
		Domain:     "global",
	}

	mockUseCase.On("AddParentRole", mock.Anything, reqBody.ChildRole, reqBody.ParentRole, "global").Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/permissions/inheritance", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
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

	w := httptest.NewRecorder()
	req := model.AssignRoleRequest{
		UserID: "u1",
		Role:   "role:admin",
		Domain: "global",
	}
	body, _ := json.Marshal(req)
	reqHTTP := httptest.NewRequest(http.MethodPost, "/permission/assign-role", bytes.NewBuffer(body))
	reqHTTP = reqHTTP.WithContext(delivery.SetContextValue(reqHTTP.Context(), delivery.UserIDKey, "u1"))

	mockUC.On("AssignRoleToUser", mock.Anything, "u1", "role:admin", "global").Return(nil)

	controller.HTTPAssignRole(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_RevokeRole(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	w := httptest.NewRecorder()
	req := model.AssignRoleRequest{
		UserID: "u1",
		Role:   "role:admin",
		Domain: "global",
	}
	body, _ := json.Marshal(req)
	reqHTTP := httptest.NewRequest(http.MethodDelete, "/permission/revoke-role", bytes.NewBuffer(body))

	mockUC.On("RevokeRoleFromUser", reqHTTP.Context(), "u1", "role:admin", "global").Return(nil)

	controller.HTTPRevokeRole(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_BatchCheck(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	w := httptest.NewRecorder()
	req := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/api/v1/users", Action: "GET"},
		},
	}
	body, err := json.Marshal(req)
	require.NoError(t, err)
	reqHTTP := httptest.NewRequest(http.MethodPost, "/permission/batch-check", bytes.NewBuffer(body))
	// Simulate middleware setting user_id
	reqHTTP = reqHTTP.WithContext(delivery.SetContextValue(reqHTTP.Context(), delivery.UserIDKey, "u1"))

	results := map[string]bool{"/api/v1/users:GET": true}
	mockUC.On("BatchCheckPermission", mock.Anything, "u1", req.Items).Return(results, nil)

	controller.HTTPBatchCheck(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response body
	var resp map[string]interface{}
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

	w := httptest.NewRecorder()
	req := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/api/v1/test", Action: "GET"},
		},
	}
	body, _ := json.Marshal(req)
	reqHTTP := httptest.NewRequest(http.MethodPost, "/permission/batch-check", bytes.NewBuffer(body))
	// NO user_id set

	controller.HTTPBatchCheck(w, reqHTTP)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPermissionController_GetAllPermissions_FiltersTenantDomain(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	w := httptest.NewRecorder()
	reqHTTP := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	ctx := delivery.SetContextValue(reqHTTP.Context(), delivery.OrganizationIDKey, "org-123")
	ctx = database.SetOrganizationContext(ctx, "org-123")
	reqHTTP = reqHTTP.WithContext(ctx)

	mockUC.On("GetAllPermissions", reqHTTP.Context()).Return([][]string{
		{"role:admin", "global", "/api/v1/users", "GET"},
		{"role:admin", "org-123", "/api/v1/projects", "GET"},
	}, nil).Once()

	controller.HTTPGetAllPermissions(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "\"global\"")
	assert.Contains(t, w.Body.String(), "\"org-123\"")
}

func TestPermissionController_GetUsersForRole_UsesResolvedTenantDomain(t *testing.T) {
	mockUC, controller := setupPermissionControllerTest()

	w := httptest.NewRecorder()
	reqHTTP := httptest.NewRequest(http.MethodGet, "/permissions/roles/role:admin/users?domain=global", nil)
	reqHTTP.SetPathValue("role", "role:admin")
	ctx := delivery.SetContextValue(reqHTTP.Context(), delivery.OrganizationIDKey, "org-123")
	ctx = database.SetOrganizationContext(ctx, "org-123")
	reqHTTP = reqHTTP.WithContext(ctx)

	mockUC.On("GetUsersForRole", reqHTTP.Context(), "role:admin", "org-123").Return([]string{"u1"}, nil).Once()

	controller.HTTPGetUsersForRole(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}
