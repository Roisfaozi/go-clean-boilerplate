package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apiKeyHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiKeyController_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUseCase := new(mocks.MockApiKeyUseCase)
	log := logrus.New()
	val := validator.New()
	controller := apiKeyHttp.NewApiKeyController(mockUseCase, log, val)

	t.Run("Success", func(t *testing.T) {
		r := gin.New()
		r.POST("/api-keys", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Set("organization_id", "org-1")
			controller.Create(c)
		})

		reqPayload := model.CreateApiKeyRequest{
			Name: "Production Key",
		}
		resPayload := &model.CreateApiKeyResponse{
			ApiKeyResponse: model.ApiKeyResponse{ID: "key-1", Name: "Production Key"},
			Key:            "sk_live_abc123",
		}

		mockUseCase.On("Create", mock.Anything, "user-1", "org-1", mock.Anything).Return(resPayload, nil)

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest("POST", "/api-keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		r := gin.New()
		r.POST("/api-keys", controller.Create)

		reqPayload := model.CreateApiKeyRequest{
			Name: "ab", // Too short
		}

		body, _ := json.Marshal(reqPayload)
		req, _ := http.NewRequest("POST", "/api-keys", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestApiKeyController_List(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUseCase := new(mocks.MockApiKeyUseCase)
	log := logrus.New()
	val := validator.New()
	controller := apiKeyHttp.NewApiKeyController(mockUseCase, log, val)

	t.Run("Success", func(t *testing.T) {
		r := gin.New()
		r.GET("/api-keys", func(c *gin.Context) {
			c.Set("organization_id", "org-1")
			controller.List(c)
		})

		resPayload := []model.ApiKeyResponse{
			{ID: "key-1", Name: "Key 1"},
		}

		mockUseCase.On("List", mock.Anything, "org-1").Return(resPayload, nil)

		req, _ := http.NewRequest("GET", "/api-keys", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}

func TestApiKeyController_Revoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUseCase := new(mocks.MockApiKeyUseCase)
	log := logrus.New()
	val := validator.New()
	controller := apiKeyHttp.NewApiKeyController(mockUseCase, log, val)

	t.Run("Success", func(t *testing.T) {
		r := gin.New()
		r.DELETE("/api-keys/:id", func(c *gin.Context) {
			c.Set("organization_id", "org-1")
			controller.Revoke(c)
		})

		mockUseCase.On("Revoke", mock.Anything, "org-1", "key-1").Return(nil)

		req, _ := http.NewRequest("DELETE", "/api-keys/key-1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUseCase.AssertExpectations(t)
	})
}
