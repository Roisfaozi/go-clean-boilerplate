package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	// If no origins are configured, we return a simple middleware that passes through
	// without adding permissive CORS headers. This relies on the browser to block
	// cross-origin requests when no CORS headers are present (Safe by Default).
	// We avoid passing an empty list to cors.New because gin-contrib/cors defaults to "*"
	if len(allowedOrigins) == 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Security: If wildcard is used, AllowCredentials must be false
	allowCredentials := true
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}

	// Security: If allowedOrigins contains wildcard "*", AllowCredentials MUST be false
	// to prevent security misconfigurations.
	allowCredentials := true
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
