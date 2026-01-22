package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	// Security: If allowed origins contains wildcard "*", AllowCredentials MUST be false
	// to prevent security misconfigurations (CORS wildcard + credentials is not allowed by browsers,
	// but explicit enforcement prevents potential server-side reflection vulnerabilities).
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
