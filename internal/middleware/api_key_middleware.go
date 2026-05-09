package middleware

import (
	"errors"
	"strings"

	apiKeyUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	authMethodContextKey   = "auth_method"
	authMethodAPIKey       = "api_key"
	authMethodJWT          = "jwt"
	apiKeyIDContextKey     = "api_key_id"
	apiKeyScopesContextKey = "api_key_scopes"
)

type APIKeyMiddleware struct {
	ApiKeyUseCase apiKeyUsecase.ApiKeyUseCase
	UserRepo      userRepository.UserRepository
	Log           *logrus.Logger
}

func NewAPIKeyMiddleware(apiKeyUseCase apiKeyUsecase.ApiKeyUseCase, userRepo userRepository.UserRepository, log *logrus.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		ApiKeyUseCase: apiKeyUseCase,
		UserRepo:      userRepo,
		Log:           log,
	}
}

func (m *APIKeyMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next() // Allow other auth methods (JWT) to handle it
			return
		}

		identity, err := m.ApiKeyUseCase.Authenticate(c.Request.Context(), apiKey)
		if err != nil {
			m.Log.WithError(err).Warn("API Key authentication failed")
			response.Unauthorized(c, err, "unauthorized")
			c.Abort()
			return
		}

		// Inject into context
		c.Set("user_id", identity.UserID)
		c.Set("organization_id", identity.OrganizationID)
		c.Set("username", identity.Username)
		c.Set(authMethodContextKey, authMethodAPIKey)
		c.Set(apiKeyIDContextKey, identity.ApiKeyID)
		c.Set(apiKeyScopesContextKey, identity.Scopes)

		c.Next()
	}
}

func (m *APIKeyMiddleware) RequireUserSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAPIKeyAuth(c) {
			response.Forbidden(c, errors.New("api key authentication is not allowed for this endpoint"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *APIKeyMiddleware) RequireScopes(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAPIKeyAuth(c) {
			c.Next()
			return
		}

		scopes, ok := GetAPIKeyScopesFromContext(c)
		if !ok || !hasRequiredScope(scopes, requiredScopes...) {
			response.Forbidden(c, errors.New("api key scope is not sufficient for this endpoint"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllScopes enforces that an API key has ALL specified scopes (AND logic).
// Use this for sensitive endpoints that require multiple permissions.
func (m *APIKeyMiddleware) RequireAllScopes(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAPIKeyAuth(c) {
			c.Next()
			return
		}

		scopes, ok := GetAPIKeyScopesFromContext(c)
		if !ok {
			response.Forbidden(c, errors.New("api key scopes not found"), "forbidden")
			c.Abort()
			return
		}

		for _, required := range requiredScopes {
			if !hasRequiredScope(scopes, required) {
				response.Forbidden(c, errors.New("api key missing required scope: "+required), "forbidden")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ScopeFromMethod returns a standard scope action suffix based on the HTTP method.
// GET/HEAD → "view", POST → "create", PUT/PATCH → "update", DELETE → "delete".
// Usage: apiKeyMiddleware.RequireScopes("project:" + middleware.ScopeFromMethod(c))
func ScopeFromMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET", "HEAD":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "view"
	}
}

func IsAPIKeyAuth(c *gin.Context) bool {
	authMethod, exists := c.Get(authMethodContextKey)
	if !exists {
		return false
	}

	authMethodStr, ok := authMethod.(string)
	return ok && authMethodStr == authMethodAPIKey
}

func GetAPIKeyScopesFromContext(c *gin.Context) ([]string, bool) {
	rawScopes, exists := c.Get(apiKeyScopesContextKey)
	if !exists {
		return nil, false
	}

	scopes, ok := rawScopes.([]string)
	if !ok {
		return nil, false
	}

	return scopes, true
}

func hasRequiredScope(grantedScopes []string, requiredScopes ...string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	for _, granted := range grantedScopes {
		if granted == "*" {
			return true
		}

		for _, required := range requiredScopes {
			if granted == required {
				return true
			}

			if strings.HasSuffix(granted, ":*") {
				prefix := strings.TrimSuffix(granted, "*")
				if strings.HasPrefix(required, prefix) {
					return true
				}
			}
		}
	}

	return false
}
