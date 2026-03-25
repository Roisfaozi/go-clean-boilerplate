package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	webhookHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupController() (*gin.Engine, *mocks.MockWebhookUseCase) {
	gin.SetMode(gin.TestMode)
	uc := new(mocks.MockWebhookUseCase)
	controller := webhookHttp.NewWebhookController(uc)

	r := gin.Default()
	r.POST("/webhooks", controller.Create)
	r.PUT("/webhooks/:id", controller.Update)
	r.DELETE("/webhooks/:id", controller.Delete)
	r.GET("/webhooks/:id", controller.FindByID)
	r.GET("/webhooks", controller.FindByOrganization)
	r.GET("/webhooks/:id/logs", controller.GetLogs)

	return r, uc
}

func TestWebhookController_Create_Success(t *testing.T) {
	r, uc := setupController()

	reqPayload := model.CreateWebhookRequest{
		Name:           "Test Webhook",
		OrganizationID: "org-1",
		URL:            "http://example.com",
		Events:         []string{"user.created"},
		Secret:         "secret123",
	}

	uc.On("Create", mock.Anything, reqPayload).Return(&model.WebhookResponse{
		ID:   "wh-1",
		Name: "Test Webhook",
	}, nil)

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_Create_BindError(t *testing.T) {
	r, _ := setupController()

	req, _ := http.NewRequest("POST", "/webhooks", bytes.NewBuffer([]byte(`invalid json`)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_Create_UseCaseError(t *testing.T) {
	r, uc := setupController()

	reqPayload := model.CreateWebhookRequest{
		Name:           "Test Webhook",
		OrganizationID: "org-1",
		URL:            "http://example.com",
		Events:         []string{"user.created"},
		Secret:         "secret123",
	}

	uc.On("Create", mock.Anything, reqPayload).Return(nil, errors.New("usecase error"))

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Since we are mocking usecase and return general error, it will map to 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_Update_Success(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	reqPayload := model.UpdateWebhookRequest{
		Name: ptrStr("Updated Webhook"),
	}

	uc.On("Update", mock.Anything, webhookID, orgID, reqPayload).Return(&model.WebhookResponse{
		ID:   webhookID,
		Name: "Updated Webhook",
	}, nil)

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/webhooks/"+webhookID+"?organization_id="+orgID, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_Update_MissingOrg(t *testing.T) {
	r, _ := setupController()

	webhookID := "wh-1"

	reqPayload := model.UpdateWebhookRequest{
		Name: ptrStr("Updated Webhook"),
	}

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/webhooks/"+webhookID, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_Update_BindError(t *testing.T) {
	r, _ := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	req, _ := http.NewRequest("PUT", "/webhooks/"+webhookID+"?organization_id="+orgID, bytes.NewBuffer([]byte(`invalid json`)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_Update_UseCaseError(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	reqPayload := model.UpdateWebhookRequest{
		Name: ptrStr("Updated Webhook"),
	}

	uc.On("Update", mock.Anything, webhookID, orgID, reqPayload).Return(nil, errors.New("not found"))

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest("PUT", "/webhooks/"+webhookID+"?organization_id="+orgID, bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_Delete_Success(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	uc.On("Delete", mock.Anything, webhookID, orgID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/webhooks/"+webhookID+"?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_Delete_MissingOrg(t *testing.T) {
	r, _ := setupController()

	webhookID := "wh-1"

	req, _ := http.NewRequest("DELETE", "/webhooks/"+webhookID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_Delete_UseCaseError(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	uc.On("Delete", mock.Anything, webhookID, orgID).Return(errors.New("not found"))

	req, _ := http.NewRequest("DELETE", "/webhooks/"+webhookID+"?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_FindByID_Success(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	uc.On("FindByID", mock.Anything, webhookID, orgID).Return(&model.WebhookResponse{
		ID:   webhookID,
		Name: "Test Webhook",
	}, nil)

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID+"?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_FindByID_MissingOrg(t *testing.T) {
	r, _ := setupController()

	webhookID := "wh-1"

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_FindByID_UseCaseError(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"

	uc.On("FindByID", mock.Anything, webhookID, orgID).Return(nil, errors.New("not found"))

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID+"?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_FindByOrganization_Success(t *testing.T) {
	r, uc := setupController()

	orgID := "org-1"

	uc.On("FindByOrganizationID", mock.Anything, orgID).Return([]model.WebhookResponse{
		{ID: "wh-1", Name: "Test Webhook"},
	}, nil)

	req, _ := http.NewRequest("GET", "/webhooks?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_FindByOrganization_MissingOrg(t *testing.T) {
	r, _ := setupController()

	req, _ := http.NewRequest("GET", "/webhooks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_FindByOrganization_UseCaseError(t *testing.T) {
	r, uc := setupController()

	orgID := "org-1"

	var res []model.WebhookResponse
	uc.On("FindByOrganizationID", mock.Anything, orgID).Return(res, errors.New("db error"))

	req, _ := http.NewRequest("GET", "/webhooks?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_GetLogs_Success(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"
	limit := 10
	offset := 0

	uc.On("FindLogs", mock.Anything, webhookID, orgID, limit, offset).Return([]interface{}{
		map[string]interface{}{"id": "log-1"},
	}, nil)

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID+"/logs?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_GetLogs_Success_CustomLimitOffset(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"
	limit := 5
	offset := 10

	uc.On("FindLogs", mock.Anything, webhookID, orgID, limit, offset).Return([]interface{}{
		map[string]interface{}{"id": "log-1"},
	}, nil)

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID+"/logs?organization_id="+orgID+"&limit=5&offset=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertExpectations(t)
}

func TestWebhookController_GetLogs_MissingOrg(t *testing.T) {
	r, _ := setupController()

	webhookID := "wh-1"

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID+"/logs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookController_GetLogs_UseCaseError(t *testing.T) {
	r, uc := setupController()

	webhookID := "wh-1"
	orgID := "org-1"
	limit := 10
	offset := 0

	var res []interface{}
	uc.On("FindLogs", mock.Anything, webhookID, orgID, limit, offset).Return(res, errors.New("not found"))

	req, _ := http.NewRequest("GET", "/webhooks/"+webhookID+"/logs?organization_id="+orgID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	uc.AssertExpectations(t)
}
