package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	orgHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type controllerTestDeps struct {
	OrgUseCase    *mocks.MockOrganizationUseCase
	MemberUseCase *mocks.MockOrganizationMemberUseCase
	Controller    *orgHttp.OrganizationController
}

func setupControllerTest() *controllerTestDeps {
	deps := &controllerTestDeps{
		OrgUseCase:    new(mocks.MockOrganizationUseCase),
		MemberUseCase: new(mocks.MockOrganizationMemberUseCase),
	}

	log := logrus.New()
	log.SetOutput(bytes.NewBuffer(nil)) // Discard logs

	validate := validator.New()
	if err := validation.RegisterCustomValidations(validate); err != nil {
		panic(err)
	}

	deps.Controller = orgHttp.NewOrganizationController(deps.OrgUseCase, deps.MemberUseCase, log, validate)
	return deps
}

func newAuthedOrgRequest(method, target string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewBuffer(body))
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	return req.WithContext(delivery.SetContextValue(req.Context(), delivery.UserIDKey, "user-123"))
}

func TestOrganizationController_CreateOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org 1", Slug: "org-1"}
		res := &model.OrganizationResponse{ID: "org-1", Name: "Org 1"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(res, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations", body)
		deps.Controller.HTTPCreateOrganization(w, reqHttp)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Success Without Slug", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org Without Slug"}
		res := &model.OrganizationResponse{ID: "org-1", Name: req.Name, Slug: "org-without-slug"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(res, nil)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHTTP := newAuthedOrgRequest(http.MethodPost, "/organizations", body)
		deps.Controller.HTTPCreateOrganization(w, reqHTTP)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Conflict", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org 1", Slug: "org-1"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(nil, exception.ErrConflict)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations", body)
		deps.Controller.HTTPCreateOrganization(w, reqHttp)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Internal Error", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.CreateOrganizationRequest{Name: "Org 1", Slug: "org-1"}

		deps.OrgUseCase.On("CreateOrganization", mock.Anything, "user-123", &req).Return(nil, exception.ErrInternalServer)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations", body)
		deps.Controller.HTTPCreateOrganization(w, reqHttp)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestOrganizationController_GetOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.OrganizationResponse{ID: "org-1", Name: "Org 1"}

		deps.OrgUseCase.On("GetOrganization", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/org-1", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPGetOrganization(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("GetOrganization", mock.Anything, "org-1").Return(nil, exception.ErrNotFound)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/org-1", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPGetOrganization(w, reqHttp)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestOrganizationController_GetOrganizationBySlug(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.OrganizationResponse{ID: "org-1", Slug: "slug-1"}

		deps.OrgUseCase.On("GetOrganizationBySlug", mock.Anything, "slug-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/slug/slug-1", nil)
		reqHttp.SetPathValue("slug", "slug-1")
		deps.Controller.HTTPGetOrganizationBySlug(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("GetOrganizationBySlug", mock.Anything, "slug-1").Return(nil, exception.ErrNotFound)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/slug/slug-1", nil)
		reqHttp.SetPathValue("slug", "slug-1")
		deps.Controller.HTTPGetOrganizationBySlug(w, reqHttp)

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
		reqHttp := newAuthedOrgRequest(http.MethodPut, "/organizations/org-1", body)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPUpdateOrganization(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.UpdateOrganizationRequest{Name: "New Name"}
		deps.OrgUseCase.On("UpdateOrganization", mock.Anything, "org-1", &req).Return(nil, exception.ErrNotFound)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPut, "/organizations/org-1", body)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPUpdateOrganization(w, reqHttp)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestOrganizationController_DeleteOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("DeleteOrganization", mock.Anything, "org-1", "user-123").Return(nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodDelete, "/organizations/org-1", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPDeleteOrganization(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("DeleteOrganization", mock.Anything, "org-1", "user-123").Return(exception.ErrForbidden)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodDelete, "/organizations/org-1", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPDeleteOrganization(w, reqHttp)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestOrganizationController_RestoreOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.OrganizationResponse{ID: "org-1", Name: "Org 1"}
		deps.OrgUseCase.On("RestoreOrganization", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations/org-1/restore", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPRestoreOrganization(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("RestoreOrganization", mock.Anything, "org-1").Return(nil, exception.ErrForbidden)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations/org-1/restore", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPRestoreOrganization(w, reqHttp)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestOrganizationController_HardDeleteOrganization(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("HardDeleteOrganization", mock.Anything, "org-1").Return(nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodDelete, "/organizations/org-1/hard", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPHardDeleteOrganization(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Bad Request", func(t *testing.T) {
		deps := setupControllerTest()
		deps.OrgUseCase.On("HardDeleteOrganization", mock.Anything, "org-1").Return(exception.ErrBadRequest)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodDelete, "/organizations/org-1/hard", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPHardDeleteOrganization(w, reqHttp)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOrganizationController_GetMyOrganizations(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := &model.UserOrganizationsResponse{Total: 1}

		deps.OrgUseCase.On("GetUserOrganizations", mock.Anything, "user-123").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/me", nil)
		deps.Controller.HTTPGetMyOrganizations(w, reqHttp)

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
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations/org-1/members", body)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPInviteMember(w, reqHttp)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Conflict", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.InviteMemberRequest{Email: "test@example.com", RoleID: "role"}
		deps.MemberUseCase.On("InviteMember", mock.Anything, "org-1", &req).Return(nil, exception.ErrConflict)

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations/org-1/members", body)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPInviteMember(w, reqHttp)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestOrganizationController_GetMembers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := []model.MemberResponse{{UserID: "u1"}}

		deps.MemberUseCase.On("GetMembers", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/org-1/members", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPGetMembers(w, reqHttp)

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
		reqHttp := newAuthedOrgRequest(http.MethodPatch, "/organizations/org-1/members/user-1", body)
		reqHttp.SetPathValue("id", "org-1")
		reqHttp.SetPathValue("userId", "user-1")
		deps.Controller.HTTPUpdateMemberRole(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestOrganizationController_RemoveMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		deps.MemberUseCase.On("RemoveMember", mock.Anything, "org-1", "user-1").Return(nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodDelete, "/organizations/org-1/members/user-1", nil)
		reqHttp.SetPathValue("id", "org-1")
		reqHttp.SetPathValue("userId", "user-1")
		deps.Controller.HTTPRemoveMember(w, reqHttp)

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
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations/invitations/accept", body)
		deps.Controller.HTTPAcceptInvitation(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		deps := setupControllerTest()
		req := model.AcceptInvitationRequest{} // Empty token

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodPost, "/organizations/invitations/accept", body)
		deps.Controller.HTTPAcceptInvitation(w, reqHttp)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOrganizationController_GetPresence(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps := setupControllerTest()
		res := []interface{}{}

		deps.MemberUseCase.On("GetPresence", mock.Anything, "org-1").Return(res, nil)

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/org-1/presence", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPGetPresence(w, reqHttp)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		deps := setupControllerTest()
		deps.MemberUseCase.On("GetPresence", mock.Anything, "org-1").Return(nil, errors.New("error"))

		w := httptest.NewRecorder()
		reqHttp := newAuthedOrgRequest(http.MethodGet, "/organizations/org-1/presence", nil)
		reqHttp.SetPathValue("id", "org-1")
		deps.Controller.HTTPGetPresence(w, reqHttp)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
