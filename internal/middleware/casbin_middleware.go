package middleware

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CasbinEnforcer defines the interface required by the middleware.
// *casbin.Enforcer satisfies this interface.
type CasbinEnforcer interface {
	Enforce(rvals ...interface{}) (bool, error)
}

// CasbinMiddleware creates a middleware for role-based authorization using Casbin.
// This middleware must be placed AFTER the JWT AuthMiddleware.
func CasbinMiddleware(enforcer CasbinEnforcer, log *logrus.Logger) gin.HandlerFunc {
	if enforcer == nil {
		log.Error("CRITICAL SECURITY ERROR: Casbin enforcer is nil. Set CASBIN_ENABLED=true.")
		if gin.Mode() == gin.ReleaseMode {
			panic("Casbin authorization cannot be disabled in production mode.")
		}
	}

	return func(c *gin.Context) {
		if enforcer == nil {
			response.ErrorResponse(c, http.StatusForbidden, errors.New("authorization unavailable"), "forbidden")
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			log.Error("Casbin middleware: user identity not found in context (AuthMiddleware missing?)")
			response.Unauthorized(c, errors.New("user not authenticated"), "unauthorized")
			c.Abort()
			return
		}

		obj := c.Request.URL.Path
		// Strip trailing slash for consistency in Casbin enforcement
		if len(obj) > 1 && obj[len(obj)-1] == '/' {
			obj = obj[:len(obj)-1]
		}
		act := c.Request.Method

		// Get organization ID for multi-tenancy (domain in Casbin)
		dom := defaultCasbinDomain
		if orgID, exists := c.Get(contextKeyOrganizationID); exists {
			if idStr, ok := orgID.(string); ok && idStr != "" {
				dom = idStr
			}
		}

		ok, err := enforcer.Enforce(userID.(string), dom, obj, act)
		if err != nil {
			log.WithError(err).Error("Casbin enforce error")
			response.InternalServerError(c, errors.New("authorization error"), "internal server error")
			c.Abort()
			return
		}

		if !ok {
			log.Warnf("Casbin authorization failed for subject '%s' in domain '%s' on %s %s", userID, dom, act, obj)
			response.Forbidden(c, errors.New("you don't have permission to access this resource"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
