package middleware

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CasbinMiddleware creates a middleware for role-based authorization using Casbin.
// This middleware must be placed AFTER the JWT AuthMiddleware.
func CasbinMiddleware(enforcer *casbin.Enforcer, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if enforcer == nil {
			c.Next()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			log.Error("Casbin middleware: user_id not found in context")
			response.Unauthorized(c, errors.New("user not authenticated"), "unauthorized")
			c.Abort()
			return
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
			log.Errorf("Casbin authorization failed for user '%s' on %s %s", userID, act, obj)
			response.Forbidden(c, errors.New("you don't have permission to access this resource"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
