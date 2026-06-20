package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	apiKeyMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
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

	t.Run("Require User Session allows JWT auth", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "jwt")
			c.Set("user_id", "user-123")
			c.Next()
		})
		r.Use(mw.RequireUserSession())
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Require All Scopes passes when all present", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"project:view", "project:manage"})
			c.Next()
		})
		r.Use(mw.RequireAllScopes("project:view", "project:manage"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Require All Scopes denies when one is missing", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"project:view"})
			c.Next()
		})
		r.Use(mw.RequireAllScopes("project:view", "project:manage"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Global wildcard scope grants access to everything", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"*"})
			c.Next()
		})
		r.Use(mw.RequireScopes("anything:here"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authenticated group chain denies valid API key because user session is required", func(t *testing.T) {
		r := gin.New()
		authUseCase := new(authMocks.MockAuthUseCase)
		authMiddleware := NewAuthMiddleware(authUseCase, log, nil)

		key := "sk_live_session_forbidden"
		identity := &apiKeyModel.ApiKeyIdentity{
			ApiKeyID:       "key-session-forbidden",
			UserID:         "user-123",
			OrganizationID: "org-456",
			Username:       "api_user",
			Scopes:         []string{"user:view"},
		}
		mockUseCase.On("Authenticate", mock.Anything, key).Return(identity, nil)

		r.Use(mw.Authenticate())
		r.Use(authMiddleware.ValidateToken())
		r.Use(mw.RequireScopeAuto())
		r.Use(mw.RequireUserSession())
		r.GET("/api/v1/users/me", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		authUseCase.AssertNotCalled(t, "ValidateAccessToken")
		authUseCase.AssertNotCalled(t, "Verify")
	})

	t.Run("Require Scope Auto derives singular scope for nested resource path", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"user:view"})
			c.Next()
		})
		r.Use(mw.RequireScopeAuto())
		r.GET("/api/v1/users/me", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Require Scope Auto denies nested resource path when derived scope missing", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"project:view"})
			c.Next()
		})
		r.Use(mw.RequireScopeAuto())
		r.GET("/api/v1/users/me", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Authorized group chain allows admin manage API key without organization context", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"admin:manage"})
			c.Next()
		})
		r.Use(mw.RequireScopes("admin:manage"))
		r.GET("/api/v1/users", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authorized group chain denies non admin API key", func(t *testing.T) {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("auth_method", "api_key")
			c.Set("api_key_scopes", []string{"user:view"})
			c.Next()
		})
		r.Use(mw.RequireScopes("admin:manage"))
		r.GET("/api/v1/users", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestScopeFromMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{"GET", "view"},
		{"HEAD", "view"},
		{"POST", "create"},
		{"PUT", "update"},
		{"PATCH", "update"},
		{"DELETE", "delete"},
		{"OPTIONS", "view"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			assert.Equal(t, tt.expected, ScopeFromMethod(tt.method))
		})
	}
}

func TestRequiredScopeFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		expected string
		ok       bool
	}{
		{name: "nested me path", path: "/api/v1/users/me", method: http.MethodGet, expected: "user:view", ok: true},
		{name: "nested avatar path", path: "/api/v1/users/me/avatar", method: http.MethodPatch, expected: "user:update", ok: true},
		{name: "organization scope alias", path: "/api/v1/organizations/org-123", method: http.MethodGet, expected: "org:view", ok: true},
		{name: "trailing slash path", path: "/api/v1/projects/", method: http.MethodGet, expected: "project:view", ok: true},
		{name: "invalid short path", path: "/health", method: http.MethodGet, expected: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, ok := requiredScopeFromRequest(tt.path, tt.method)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, scope)
		})
	}
}
