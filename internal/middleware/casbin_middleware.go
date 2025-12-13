package middleware

import (
	"errors"

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
	return func(c *gin.Context) {
		if enforcer == nil {
			c.Next()
			return
		}

		// AuthMiddleware ensures user_id is set in the context
		userID, exists := c.Get("user_id")
		if !exists {
			log.Error("Casbin middleware: user identity not found in context (AuthMiddleware missing?)")
			response.Unauthorized(c, errors.New("user not authenticated"), "unauthorized")
			c.Abort()
			return
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		// The policy checks if the user (userID) has permission on obj/act.
		// Grouping policies (g) map userID to roles.
		ok, err := enforcer.Enforce(userID.(string), obj, act)
		if err != nil {
			log.WithError(err).Error("Casbin enforce error")
			response.InternalServerError(c, errors.New("authorization error"), "internal server error")
			c.Abort()
			return
		}

		if !ok {
			log.Warnf("Casbin authorization failed for subject '%s' on %s %s", userID, act, obj)
			response.Forbidden(c, errors.New("you don't have permission to access this resource"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
