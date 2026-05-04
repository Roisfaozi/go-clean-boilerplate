package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	apiKeyMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeyMiddleware_Authenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logrus.New()

	mockUseCase := new(apiKeyMocks.MockApiKeyUseCase)
	mockUserRepo := new(userMocks.MockUserRepository)

	mw := NewAPIKeyMiddleware(mockUseCase, mockUserRepo, log)

	t.Run("Valid API Key", func(t *testing.T) {
		r := gin.New()
		r.Use(mw.Authenticate())
		r.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			authMethod, _ := c.Get("auth_method")
			apiKeyID, _ := c.Get("api_key_id")
			scopes, _ := c.Get("api_key_scopes")
			c.JSON(http.StatusOK, gin.H{
				"user_id":     userID,
				"auth_method": authMethod,
				"api_key_id":  apiKeyID,
				"scopes":      scopes,
			})
		})

		key := "sk_live_valid_key"
		identity := &apiKeyModel.ApiKeyIdentity{
			ApiKeyID:       "key-123",
			UserID:         "user-123",
			OrganizationID: "org-456",
			Username:       "api_user",
			Scopes:         []string{"project:view"},
		}

		mockUseCase.On("Authenticate", mock.Anything, key).Return(identity, nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user-123")
		assert.Contains(t, w.Body.String(), "api_key")
		assert.Contains(t, w.Body.String(), "key-123")
		assert.Contains(t, w.Body.String(), "project:view")
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		r := gin.New()
		r.Use(mw.Authenticate())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "should not reach here")
		})

		key := "sk_live_invalid"
		mockUseCase.On("Authenticate", mock.Anything, key).Return(nil, assert.AnError)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Require Scopes allows JWT auth", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "jwt")
			c.Set("user_id", "user-123")
			c.Next()
		})
		r.Use(mw.RequireScopes("project:view"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Require Scopes denies API key without scope", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"project:view"})
			c.Next()
		})
		r.Use(mw.RequireScopes("project:manage"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Require Scopes allows wildcard API key scope", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"project:*"})
			c.Next()
		})
		r.Use(mw.RequireScopes("project:manage"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Require User Session denies API key auth", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Next()
		})
		r.Use(mw.RequireUserSession())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
