package test

import (
	"bytes"
	"encoding/json"
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

func TestWebhookController_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := new(mocks.MockWebhookUseCase)
	controller := webhookHttp.NewWebhookController(uc)

	r := gin.Default()
	r.POST("/webhooks", controller.Create)

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

func TestWebhookController_FindByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := new(mocks.MockWebhookUseCase)
	controller := webhookHttp.NewWebhookController(uc)

	r := gin.Default()
	r.GET("/webhooks/:id", controller.FindByID)

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
