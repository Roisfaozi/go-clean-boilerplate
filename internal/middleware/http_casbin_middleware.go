package middleware

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

// HTTPCasbinMiddleware enforces RBAC policy based on user_id, organization domain, URL path, and method.
func HTTPCasbinMiddleware(enforcer CasbinEnforcer, log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enforcer == nil {
				response.WriteError(w, http.StatusForbidden, errors.New("authorization unavailable"), "forbidden")
				return
			}

			userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, errors.New("user not authenticated"), "unauthorized")
				return
			}

			obj := r.URL.Path
			if len(obj) > 1 && obj[len(obj)-1] == '/' {
				obj = obj[:len(obj)-1]
			}
			act := r.Method

			dom := defaultCasbinDomain
			if orgID := database.GetOrganizationID(r.Context()); orgID != "" {
				dom = orgID
			}

			allowed, err := enforcer.Enforce(userID, dom, obj, act)
			if err != nil {
				if log != nil {
					log.WithError(err).Error("Casbin enforce error")
				}
				response.WriteError(w, http.StatusInternalServerError, errors.New("authorization error"), "internal server error")
				return
			}

			if !allowed {
				if log != nil {
					log.Warnf("Casbin authorization failed for subject '%s' in domain '%s' on %s %s", userID, dom, act, obj)
				}
				response.WriteError(w, http.StatusForbidden, errors.New("you don't have permission to access this resource"), "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
