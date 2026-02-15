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

	// Check for wildcard
	allowAllOrigins := true
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAllOrigins = false
			break
		}
	}

	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Organization-ID", "X-Organization-Slug"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if allowAllOrigins {
		config.AllowOriginFunc = func(origin string) bool {
			return true
		}
	} else {
		config.AllowOrigins = allowedOrigins
	}

	return cors.New(config)
}
