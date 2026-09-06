package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
)

// HTTPValidateToken verifies Bearer token or cookie on incoming HTTP requests.
func (m *AuthMiddleware) HTTPValidateToken() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if already authenticated via API Key
			if _, isAPIKey := delivery.GetContextString(r.Context(), delivery.APIKeyIDKey); isAPIKey {
				next.ServeHTTP(w, r)
				return
			}

			token := ""
			authHeader := r.Header.Get(headerAuthorization)
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == authSchemeBearer {
					token = parts[1]
				}
			}

			if token == "" {
				if cookie, err := r.Cookie(cookieAccessToken); err == nil && cookie.Value != "" {
					token = cookie.Value
				}
			}

			if token == "" {
				response.WriteError(w, http.StatusUnauthorized, errors.New("token is required"), "unauthorized")
				return
			}

			claims, err := m.AuthUseCase.ValidateAccessToken(token)
			if err != nil {
				if m.Log != nil {
					m.Log.WithError(err).Warn("Token validation failed")
				}
				response.WriteError(w, http.StatusUnauthorized, err, "unauthorized")
				return
			}

			session, err := m.AuthUseCase.Verify(r.Context(), claims.UserID, claims.SessionID)
			if err != nil {
				if m.Log != nil {
					m.Log.WithError(err).Warn("Session verification failed with database/redis error")
				}
				response.WriteError(w, http.StatusInternalServerError, errors.New("could not verify session"), "internal server error")
				return
			}
			if session == nil {
				if m.Log != nil {
					m.Log.Warn("Session is not valid or has been revoked")
				}
				response.WriteError(w, http.StatusUnauthorized, errors.New("invalid or expired session"), "unauthorized")
				return
			}

			ctx := authcontext.WithUserID(r.Context(), claims.UserID)
			ctx = authcontext.WithSessionID(ctx, claims.SessionID)
			ctx = authcontext.WithRole(ctx, claims.Role)
			ctx = authcontext.WithUsername(ctx, claims.Username)
			ctx = delivery.SetContextValue(ctx, delivery.UserIDKey, claims.UserID)
			ctx = delivery.SetContextValue(ctx, delivery.SessionIDKey, claims.SessionID)
			ctx = delivery.SetContextValue(ctx, delivery.RoleKey, claims.Role)
			ctx = delivery.SetContextValue(ctx, delivery.UsernameKey, claims.Username)
			ctx = delivery.SetContextValue(ctx, delivery.AuthMethodKey, "jwt")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
