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

		userID, exists := c.Get("user_id")
		if !exists {
			// Try to get from x-user-role if user_id is not set but role is? 
			// No, standard is user_id -> role lookup via casbin usually, OR role directly.
			// The current implementation assumes userID is the subject in policy?
			// Wait, the test uses "role:admin" as subject.
			// If AuthMiddleware sets "x-user-role", we should probably use THAT if our policy is role-based.
			// Let's check what AuthMiddleware sets. It sets "x-user-id" and "x-user-role".
			
			// If we want RBAC based on ROLE, we should use the role from context.
			// If we want RBAC based on USER, we use user_id.
			// Standard simple RBAC usually checks ROLE.
			
			// Let's check "x-user-role" first for RBAC.
			role, roleExists := c.Get("x-user-role")
			if roleExists {
				userID = role // Use role as subject
			} else {
				// Fallback to user_id or error
				uid, idExists := c.Get("x-user-id")
				if !idExists {
					log.Error("Casbin middleware: user identity not found in context")
					response.Unauthorized(c, errors.New("user not authenticated"), "unauthorized")
					c.Abort()
					return
				}
				userID = uid
			}
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		ok, err := enforcer.Enforce(userID.(string), obj, act)
		if err != nil {
			log.WithError(err).Error("Casbin enforce error")
			response.InternalServerError(c, errors.New("authorization error"), "internal server error")
			c.Abort()
			return
		}

		if !ok {
			log.Errorf("Casbin authorization failed for subject '%s' on %s %s", userID, act, obj)
			response.Forbidden(c, errors.New("you don't have permission to access this resource"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}