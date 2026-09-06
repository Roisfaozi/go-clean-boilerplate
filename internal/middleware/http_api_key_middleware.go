package middleware

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

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
