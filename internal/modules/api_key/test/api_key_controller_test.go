package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	apiKeyHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiKeyController_Create(t *testing.T) {
	mockUseCase := new(mocks.MockApiKeyUseCase)
	log := logrus.New()
	val := validator.New()
	controller := apiKeyHttp.NewApiKeyController(mockUseCase, log, val)

	t.Run("Success", func(t *testing.T) {
		reqPayload := model.CreateApiKeyRequest{
			Name: "Production Key",
		}
		resPayload := &model.CreateApiKeyResponse{
			ApiKeyResponse: model.ApiKeyResponse{ID: "key-1", Name: "Production Key"},
			Key:            "sk_live_abc123",
		}

		mockUseCase.On("Create", mock.Anything, "user-1", "org-1", mock.Anything).Return(resPayload, nil)

		body, _ := json.Marshal(reqPayload)
		req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
		ctx := delivery.SetContextValue(req.Context(), delivery.UserIDKey, "user-1")
		ctx = database.SetOrganizationContext(ctx, "org-1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		controller.HTTPCreate(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqPayload := model.CreateApiKeyRequest{
			Name: "ab", // Too short
		}

		body, _ := json.Marshal(reqPayload)
		req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		controller.HTTPCreate(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestApiKeyController_List(t *testing.T) {
	mockUseCase := new(mocks.MockApiKeyUseCase)
	log := logrus.New()
	val := validator.New()
	controller := apiKeyHttp.NewApiKeyController(mockUseCase, log, val)

	t.Run("Success", func(t *testing.T) {
		resPayload := []model.ApiKeyResponse{
			{ID: "key-1", Name: "Key 1"},
		}

		mockUseCase.On("List", mock.Anything, "org-1").Return(resPayload, nil)

		req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
		req = req.WithContext(database.SetOrganizationContext(req.Context(), "org-1"))
		w := httptest.NewRecorder()
		controller.HTTPList(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestApiKeyController_Revoke(t *testing.T) {
	mockUseCase := new(mocks.MockApiKeyUseCase)
	log := logrus.New()
	val := validator.New()
	controller := apiKeyHttp.NewApiKeyController(mockUseCase, log, val)

	t.Run("Success", func(t *testing.T) {
		mockUseCase.On("Revoke", mock.Anything, "org-1", "key-1").Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api-keys/key-1", nil)
		req.SetPathValue("id", "key-1")
		req = req.WithContext(database.SetOrganizationContext(req.Context(), "org-1"))
		w := httptest.NewRecorder()
		controller.HTTPRevoke(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}
