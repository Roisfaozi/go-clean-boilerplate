package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	orgHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type orgControllerTestDeps struct {
	OrgUC    *mocks.MockOrganizationUseCase
	MemberUC *mocks.MockOrganizationMemberUseCase
	Ctrl     *orgHttp.OrganizationController
}

func setupOrganizationControllerTest() *orgControllerTestDeps {
	orgUC := new(mocks.MockOrganizationUseCase)
	memberUC := new(mocks.MockOrganizationMemberUseCase)
	log := logrus.New()
	log.SetOutput(bytes.NewBuffer(nil)) // Discard logs
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	ctrl := orgHttp.NewOrganizationController(orgUC, memberUC, log, v)

	return &orgControllerTestDeps{
		OrgUC:    orgUC,
		MemberUC: memberUC,
		Ctrl:     ctrl,
	}
}

func TestOrganizationController_Create_Success(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user-123")
	requestBody := model.CreateOrganizationRequest{
		Name: "Test Org",
		Slug: "test-org",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	expectedResponse := &model.OrganizationResponse{
		ID:   "org-123",
		Name: "Test Org",
		Slug: "test-org",
	}

	deps.OrgUC.On("CreateOrganization", mock.Anything, "user-123", &requestBody).Return(expectedResponse, nil)

	deps.Ctrl.CreateOrganization(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "org-123", data["id"])
	assert.Equal(t, "Test Org", data["name"])
}

func TestOrganizationController_Create_Unauthorized(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No user_id in context

	requestBody := model.CreateOrganizationRequest{
		Name: "Test Org",
		Slug: "test-org",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(jsonBody))

	deps.Ctrl.CreateOrganization(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOrganizationController_Create_InvalidBody(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "user-123")

	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer([]byte("{invalid-json}")))

	deps.Ctrl.CreateOrganization(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrganizationController_Create_Conflict(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user-123")
	requestBody := model.CreateOrganizationRequest{
		Name: "Test Org",
		Slug: "existing-slug",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	deps.OrgUC.On("CreateOrganization", mock.Anything, "user-123", &requestBody).Return(nil, exception.ErrConflict)

	deps.Ctrl.CreateOrganization(c)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestOrganizationController_Create_InternalError(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user-123")
	requestBody := model.CreateOrganizationRequest{
		Name: "Test Org",
		Slug: "test-org",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	deps.OrgUC.On("CreateOrganization", mock.Anything, "user-123", &requestBody).Return(nil, errors.New("db error"))

	deps.Ctrl.CreateOrganization(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestOrganizationController_Create_XSS(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user-123")

	requestBody := model.CreateOrganizationRequest{
		Name: "<script>alert(1)</script>",
		Slug: "xss-org",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	deps.Ctrl.CreateOrganization(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestOrganizationController_GetOrganization_Success(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/organizations/org-123", nil)

	expectedResponse := &model.OrganizationResponse{
		ID:   "org-123",
		Name: "Test Org",
	}

	deps.OrgUC.On("GetOrganization", mock.Anything, "org-123").Return(expectedResponse, nil)

	deps.Ctrl.GetOrganization(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrganizationController_GetOrganization_NotFound(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/organizations/org-123", nil)

	deps.OrgUC.On("GetOrganization", mock.Anything, "org-123").Return(nil, exception.ErrNotFound)

	deps.Ctrl.GetOrganization(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrganizationController_UpdateOrganization_Success(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	requestBody := model.UpdateOrganizationRequest{
		Name: "Updated Name",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPut, "/organizations/org-123", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	expectedResponse := &model.OrganizationResponse{
		ID:   "org-123",
		Name: "Updated Name",
	}

	deps.OrgUC.On("UpdateOrganization", mock.Anything, "org-123", &requestBody).Return(expectedResponse, nil)

	deps.Ctrl.UpdateOrganization(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrganizationController_DeleteOrganization_Success(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user-123")
	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/organizations/org-123", nil)

	deps.OrgUC.On("DeleteOrganization", mock.Anything, "org-123", "user-123").Return(nil)

	deps.Ctrl.DeleteOrganization(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOrganizationController_DeleteOrganization_Forbidden(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "user-123")
	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/organizations/org-123", nil)

	deps.OrgUC.On("DeleteOrganization", mock.Anything, "org-123", "user-123").Return(exception.ErrForbidden)

	deps.Ctrl.DeleteOrganization(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOrganizationController_InviteMember_Success(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	requestBody := model.InviteMemberRequest{
		Email:  "test@example.com",
		RoleID: "role-admin",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations/org-123/members", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	expectedResponse := &model.MemberResponse{
		ID:     "member-1",
		RoleID: "role-admin",
	}

	deps.MemberUC.On("InviteMember", mock.Anything, "org-123", &requestBody).Return(expectedResponse, nil)

	deps.Ctrl.InviteMember(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestOrganizationController_InviteMember_Conflict(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	requestBody := model.InviteMemberRequest{
		Email:  "test@example.com",
		RoleID: "role-admin",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPost, "/organizations/org-123/members", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	deps.MemberUC.On("InviteMember", mock.Anything, "org-123", &requestBody).Return(nil, exception.ErrConflict)

	deps.Ctrl.InviteMember(c)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestOrganizationController_Update_XSS(t *testing.T) {
	deps := setupOrganizationControllerTest()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: "id", Value: "org-123"}}
	requestBody := model.UpdateOrganizationRequest{
		Name: "<script>alert(1)</script>",
	}
	jsonBody, _ := json.Marshal(requestBody)
	c.Request, _ = http.NewRequest(http.MethodPut, "/organizations/org-123", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	deps.Ctrl.UpdateOrganization(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
