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

type controllerTestDeps struct {
	OrgUseCase    *mocks.MockOrganizationUseCase
	MemberUseCase *mocks.MockOrganizationMemberUseCase
	Controller    *orgHttp.OrganizationController
	Router        *gin.Engine
}

func setupControllerTest() *controllerTestDeps {
	gin.SetMode(gin.TestMode)

	deps := &controllerTestDeps{
		OrgUseCase:    new(mocks.MockOrganizationUseCase),
		MemberUseCase: new(mocks.MockOrganizationMemberUseCase),
	}

	log := logrus.New()
	log.SetOutput(bytes.NewBuffer(nil)) // Discard logs

	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)

	deps.Controller = orgHttp.NewOrganizationController(deps.OrgUseCase, deps.MemberUseCase, log, validate)

	deps.Router = gin.New()
	// Mock Auth Middleware by manually setting user_id in context
	deps.Router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})

	// Setup routes manually for testing
	deps.Router.POST("/organizations", deps.Controller.CreateOrganization)
	deps.Router.GET("/organizations/:id", deps.Controller.GetOrganization)
	deps.Router.GET("/organizations/slug/:slug", deps.Controller.GetOrganizationBySlug)
	deps.Router.PUT("/organizations/:id", deps.Controller.UpdateOrganization)
	deps.Router.DELETE("/organizations/:id", deps.Controller.DeleteOrganization)
	deps.Router.GET("/organizations/me", deps.Controller.GetMyOrganizations)

	deps.Router.POST("/organizations/invitations/accept", deps.Controller.AcceptInvitation)
	deps.Router.POST("/organizations/:id/members", deps.Controller.InviteMember)
	deps.Router.GET("/organizations/:id/members", deps.Controller.GetMembers)
	deps.Router.PATCH("/organizations/:id/members/:userId", deps.Controller.UpdateMemberRole)
	deps.Router.DELETE("/organizations/:id/members/:userId", deps.Controller.RemoveMember)
	deps.Router.GET("/organizations/:id/presence", deps.Controller.GetPresence)

	return deps
}

func TestOrganizationController_CreateOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org 1", Slug: "org-1"}
		res := &model.OrganizationResponse{ID: "org-1", Name: "Org 1"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(res, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Conflict", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org 1", Slug: "org-1"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(nil, exception.ErrConflict)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Internal Error", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org 1", Slug: "org-1"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(nil, exception.ErrInternalServer)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestOrganizationController_GetOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.OrganizationResponse{ID: "org-1", Name: "Org 1"}

		deps.OrgUseCase.On("GetOrganization", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/org-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("GetOrganization", mock.Anything, "org-1").Return(nil, exception.ErrNotFound)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/org-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestOrganizationController_GetOrganizationBySlug(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.OrganizationResponse{ID: "org-1", Slug: "slug-1"}

		deps.OrgUseCase.On("GetOrganizationBySlug", mock.Anything, "slug-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/slug/slug-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("GetOrganizationBySlug", mock.Anything, "slug-1").Return(nil, exception.ErrNotFound)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/slug/slug-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestOrganizationController_UpdateOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.UpdateOrganizationRequest{Name: "New Name"}
		res := &model.OrganizationResponse{ID: "org-1", Name: "New Name"}

		deps.OrgUseCase.On("UpdateOrganization", mock.Anything, "org-1", &req).Return(res, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("PUT", "/organizations/org-1", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.UpdateOrganizationRequest{Name: "New Name"}
		deps.OrgUseCase.On("UpdateOrganization", mock.Anything, "org-1", &req).Return(nil, exception.ErrNotFound)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("PUT", "/organizations/org-1", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestOrganizationController_DeleteOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("DeleteOrganization", mock.Anything, "org-1", "user-123").Return(nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("DELETE", "/organizations/org-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("DeleteOrganization", mock.Anything, "org-1", "user-123").Return(exception.ErrForbidden)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("DELETE", "/organizations/org-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestOrganizationController_GetMyOrganizations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.UserOrganizationsResponse{Total: 1}

		deps.OrgUseCase.On("GetUserOrganizations", mock.Anything, "user-123").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/me", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestOrganizationController_InviteMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.InviteMemberRequest{Email: "test@example.com", RoleID: "role"}
		res := &model.MemberResponse{UserID: "u1"}

		deps.MemberUseCase.On("InviteMember", mock.Anything, "org-1", &req).Return(res, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations/org-1/members", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Conflict", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.InviteMemberRequest{Email: "test@example.com", RoleID: "role"}
		deps.MemberUseCase.On("InviteMember", mock.Anything, "org-1", &req).Return(nil, exception.ErrConflict)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations/org-1/members", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestOrganizationController_GetMembers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := []model.MemberResponse{{UserID: "u1"}}

		deps.MemberUseCase.On("GetMembers", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/org-1/members", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestOrganizationController_UpdateMemberRole(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.UpdateMemberRequest{RoleID: "new-role"}
		res := &model.MemberResponse{RoleID: "new-role"}

		deps.MemberUseCase.On("UpdateMember", mock.Anything, "org-1", "user-1", &req).Return(res, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("PATCH", "/organizations/org-1/members/user-1", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestOrganizationController_RemoveMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		deps.MemberUseCase.On("RemoveMember", mock.Anything, "org-1", "user-1").Return(nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("DELETE", "/organizations/org-1/members/user-1", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestOrganizationController_AcceptInvitation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.AcceptInvitationRequest{Token: "token"}

		deps.MemberUseCase.On("AcceptInvitation", mock.Anything, &req).Return(nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations/invitations/accept", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.AcceptInvitationRequest{} // Empty token

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("POST", "/organizations/invitations/accept", bytes.NewBuffer(body))
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOrganizationController_GetPresence(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := []interface{}{}

		deps.MemberUseCase.On("GetPresence", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/org-1/presence", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		deps := setupControllerTest()
		deps.MemberUseCase.On("GetPresence", mock.Anything, "org-1").Return(nil, errors.New("error"))

		w := httptest.NewRecorder()
		reqHttp, _ := http.NewRequest("GET", "/organizations/org-1/presence", nil)
		deps.Router.ServeHTTP(w, reqHttp)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
