package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	apiKeyHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type apiKeyControllerDeps struct {
	UseCase *mocks.MockApiKeyUseCase
}

func setupApiKeyControllerTest() (*apiKeyControllerDeps, *gin.Engine, *apiKeyHttp.ApiKeyController) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	deps := &apiKeyControllerDeps{
		UseCase: new(mocks.MockApiKeyUseCase),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	validator := config.NewValidator()
	controller := apiKeyHttp.NewApiKeyController(deps.UseCase, log, validator)

	return deps, engine, controller
}

func TestApiKeyController_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.POST("/api-keys", func(c *gin.Context) {
			c.Set("user_id", "user-id")
			c.Set("organization_id", "org-id")
			controller.Create(c)
		})

		expiresAt := time.Now().Add(24 * time.Hour)
		reqBody := model.CreateApiKeyRequest{
			Name:      "Test Key",
			Scopes:    []string{"read", "write"},
			ExpiresAt: &expiresAt,
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(jsonValue))

		res := &model.CreateApiKeyResponse{
			ApiKeyResponse: model.ApiKeyResponse{
				ID:             "key-1",
				Name:           "Test Key",
				OrganizationID: "org-id",
				UserID:         "user-id",
				Scopes:         []string{"read", "write"},
				ExpiresAt:      &expiresAt,
				IsActive:       true,
			},
			Key: "sk_live_generatedkey",
		}

		deps.UseCase.EXPECT().Create(mock.Anything, "user-id", "org-id", mock.AnythingOfType("*model.CreateApiKeyRequest")).Return(res, nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.POST("/api-keys", func(c *gin.Context) {
			c.Set("user_id", "user-id")
			c.Set("organization_id", "org-id")
			controller.Create(c)
		})

		// Missing required field 'Name'
		reqBody := model.CreateApiKeyRequest{
			Scopes: []string{"read", "write"},
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(jsonValue))

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Missing Organization ID", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.POST("/api-keys", func(c *gin.Context) {
			c.Set("user_id", "user-id")
			controller.Create(c)
		})

		reqBody := model.CreateApiKeyRequest{
			Name:   "Test Key",
			Scopes: []string{"read", "write"},
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(jsonValue))

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.POST("/api-keys", func(c *gin.Context) {
			c.Set("user_id", "user-id")
			c.Set("organization_id", "org-id")
			controller.Create(c)
		})

		reqBody := model.CreateApiKeyRequest{
			Name:   "Test Key",
			Scopes: []string{"read", "write"},
		}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(jsonValue))

		deps.UseCase.EXPECT().Create(mock.Anything, "user-id", "org-id", mock.AnythingOfType("*model.CreateApiKeyRequest")).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}

func TestApiKeyController_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.GET("/api-keys", func(c *gin.Context) {
			c.Set("organization_id", "org-id")
			controller.List(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/api-keys", nil)

		res := []model.ApiKeyResponse{
			{
				ID:             "key-1",
				Name:           "Key 1",
				OrganizationID: "org-id",
			},
		}

		deps.UseCase.EXPECT().List(mock.Anything, "org-id").Return(res, nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Missing Organization ID", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.GET("/api-keys", func(c *gin.Context) {
			controller.List(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/api-keys", nil)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.GET("/api-keys", func(c *gin.Context) {
			c.Set("organization_id", "org-id")
			controller.List(c)
		})

		req, _ := http.NewRequest(http.MethodGet, "/api-keys", nil)

		deps.UseCase.EXPECT().List(mock.Anything, "org-id").Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}

func TestApiKeyController_Revoke(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.DELETE("/api-keys/:id", func(c *gin.Context) {
			c.Set("organization_id", "org-id")
			controller.Revoke(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/api-keys/key-1", nil)

		deps.UseCase.EXPECT().Revoke(mock.Anything, "org-id", "key-1").Return(nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Missing Organization ID", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.DELETE("/api-keys/:id", func(c *gin.Context) {
			controller.Revoke(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/api-keys/key-1", nil)

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, engine, controller := setupApiKeyControllerTest()

		engine.DELETE("/api-keys/:id", func(c *gin.Context) {
			c.Set("organization_id", "org-id")
			controller.Revoke(c)
		})

		req, _ := http.NewRequest(http.MethodDelete, "/api-keys/key-1", nil)

		deps.UseCase.EXPECT().Revoke(mock.Anything, "org-id", "key-1").Return(errors.New("db error")).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}