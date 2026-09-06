package middleware

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
)

// HTTPValidateWebSocketToken verifies WS ticket and injects WS identity.
func (m *AuthMiddleware) HTTPValidateWebSocketToken() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ticket := r.URL.Query().Get("ticket")
			if ticket == "" {
				response.WriteError(w, http.StatusUnauthorized, errors.New("ticket is required"), "unauthorized")
				return
			}

			userCtx, err := m.TicketManager.ValidateTicket(r.Context(), ticket)
			if err != nil {
				if m.Log != nil {
					m.Log.WithError(err).Warn("Invalid or expired WebSocket ticket")
				}
				response.WriteError(w, http.StatusUnauthorized, errors.New("invalid or expired ticket"), "unauthorized")
				return
			}

			ctx := authcontext.WithUserID(r.Context(), userCtx.UserID)
			ctx = authcontext.WithSessionID(ctx, userCtx.SessionID)
			ctx = authcontext.WithRole(ctx, userCtx.Role)
			ctx = authcontext.WithUsername(ctx, userCtx.Username)
			ctx = delivery.SetContextValue(ctx, delivery.UserIDKey, userCtx.UserID)
			ctx = delivery.SetContextValue(ctx, delivery.SessionIDKey, userCtx.SessionID)
			ctx = delivery.SetContextValue(ctx, delivery.RoleKey, userCtx.Role)
			ctx = delivery.SetContextValue(ctx, delivery.UsernameKey, userCtx.Username)
			ctx = delivery.SetContextValue(ctx, delivery.AuthMethodKey, "jwt")

			if userCtx.OrganizationID != "" {
				ctx = database.SetOrganizationContext(ctx, userCtx.OrganizationID)
				ctx = delivery.SetContextValue(ctx, delivery.OrganizationIDKey, userCtx.OrganizationID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
