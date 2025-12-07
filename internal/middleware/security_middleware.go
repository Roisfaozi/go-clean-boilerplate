package middleware

import "github.com/gin-gonic/gin"

// SecurityMiddleware adds common security headers to the response.
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Referrer-Policy is also good practice
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
