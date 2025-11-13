package middleware

import (
	"errors"
	"strings"

	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a middleware to check for a valid JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, errors.New("authorization header is required"))
			c.Abort()
			return
		}

		// Format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, errors.New("invalid authorization header format"))
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, errors.New("token is required"))
			c.Abort()
			return
		}

		// TODO: Validate the token (you can use your JWT validation logic here)
		// For now, we'll just set the user ID from the token
		// In a real application, you would decode the token and validate it
		// and then set the user information in the context
		userID := "user123" // This should come from the token
		c.Set("user_id", userID)

		c.Next()
	}
}
