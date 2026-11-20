package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	apiKeyUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

const apiKeyHeaderName = "X-API-Key"

var apiKeyScopeResourceAliases = map[string]string{
	"organization": "org",
}

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

// HTTPAuthenticate validates X-API-Key header and populates context.
func (m *APIKeyMiddleware) HTTPAuthenticate() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(apiKeyHeaderName)
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			identity, err := m.ApiKeyUseCase.Authenticate(r.Context(), apiKey)
			if err != nil {
				if m.Log != nil {
					m.Log.WithError(err).Warn("API Key authentication failed")
				}
				response.WriteError(w, http.StatusUnauthorized, err, "unauthorized")
				return
			}

			ctx := r.Context()
			ctx = authcontext.WithUserID(ctx, identity.UserID)
			ctx = authcontext.WithUsername(ctx, identity.Username)
			ctx = database.SetOrganizationContext(ctx, identity.OrganizationID)
			ctx = delivery.SetContextValue(ctx, delivery.UserIDKey, identity.UserID)
			ctx = delivery.SetContextValue(ctx, delivery.OrganizationIDKey, identity.OrganizationID)
			ctx = delivery.SetContextValue(ctx, delivery.UsernameKey, identity.Username)
			ctx = delivery.SetContextValue(ctx, delivery.AuthMethodKey, "api_key")
			ctx = delivery.SetContextValue(ctx, delivery.APIKeyIDKey, identity.ApiKeyID)
			ctx = delivery.SetContextValue(ctx, delivery.ScopesKey, identity.Scopes)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HTTPRequireScopes ensures request has at least one of the required scopes if authenticated via API key.
func (m *APIKeyMiddleware) HTTPRequireScopes(requiredScopes ...string) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authMethod, _ := delivery.GetContextString(r.Context(), delivery.AuthMethodKey)
			if authMethod != "api_key" {
				next.ServeHTTP(w, r)
				return
			}

			scopesVal, _ := r.Context().Value(delivery.ScopesKey).([]string)
			for _, required := range requiredScopes {
				if hasRequiredScope(scopesVal, required) {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.WriteError(w, http.StatusForbidden, errors.New("insufficient permissions: required scope missing"), "forbidden")
		})
	}
}

// HTTPRequireScopeAuto derives the required scope from the URL path and HTTP method.
func (m *APIKeyMiddleware) HTTPRequireScopeAuto() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authMethod, _ := delivery.GetContextString(r.Context(), delivery.AuthMethodKey)
			if authMethod != "api_key" {
				next.ServeHTTP(w, r)
				return
			}

			requiredScope, ok := requiredScopeFromRequest(r.URL.Path, r.Method)
			if !ok {
				if m.Log != nil {
					m.Log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"method": r.Method,
					}).Warn("API Key scope enforcement failed: undeterminable required scope")
				}
				response.WriteError(w, http.StatusForbidden, errors.New("api key scope required"), "forbidden")
				return
			}

			scopesVal, _ := r.Context().Value(delivery.ScopesKey).([]string)
			if !hasRequiredScope(scopesVal, requiredScope) {
				if m.Log != nil {
					m.Log.WithFields(logrus.Fields{
						"required_scope": requiredScope,
						"granted_scopes": scopesVal,
						"path":           r.URL.Path,
					}).Warn("API Key scope enforcement failed (Auto)")
				}
				response.WriteError(w, http.StatusForbidden, errors.New("api key scope is not sufficient: "+requiredScope), "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPRequireUserSession blocks API key authentication from session-only endpoints.
func (m *APIKeyMiddleware) HTTPRequireUserSession() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authMethod, _ := delivery.GetContextString(r.Context(), delivery.AuthMethodKey)
			if authMethod == "api_key" {
				response.WriteError(w, http.StatusForbidden, errors.New("api key authentication is not allowed for this endpoint"), "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func requiredScopeFromRequest(path, method string) (string, bool) {
	resource, ok := resourceFromRequestPath(path)
	if !ok {
		return "", false
	}

	return resource + ":" + ScopeFromMethod(method), true
}

func resourceFromRequestPath(path string) (string, bool) {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 3 {
		return "", false
	}

	resource := pathParts[2]
	if resource == "" {
		return "", false
	}

	if strings.HasSuffix(resource, "s") && !strings.HasSuffix(resource, "ss") {
		resource = resource[:len(resource)-1]
	}

	if alias, ok := apiKeyScopeResourceAliases[resource]; ok {
		resource = alias
	}

	return resource, true
}

// ScopeFromMethod returns a standard scope action suffix based on the HTTP method.
// GET/HEAD → "view", POST → "create", PUT/PATCH → "update", DELETE → "delete".
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

func hasRequiredScope(grantedScopes []string, requiredScopes ...string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	for _, granted := range grantedScopes {
		if granted == "*" {
			return true
		}

		grantedParts := strings.Split(granted, ":")
		if len(grantedParts) != 2 {
			continue
		}
		grantedResource := grantedParts[0]
		grantedAction := grantedParts[1]

		for _, required := range requiredScopes {
			if granted == required {
				return true
			}

			requiredParts := strings.Split(required, ":")
			if len(requiredParts) != 2 {
				continue
			}
			requiredResource := requiredParts[0]
			requiredAction := requiredParts[1]

			if grantedResource == requiredResource || grantedResource == "*" {
				// Manage covers everything
				if grantedAction == "manage" || grantedAction == "*" {
					return true
				}
				// View covers read/view
				if (grantedAction == "view" || grantedAction == "read") &&
					(requiredAction == "view" || requiredAction == "read") {
					return true
				}
			}
		}
	}

	return false
}
