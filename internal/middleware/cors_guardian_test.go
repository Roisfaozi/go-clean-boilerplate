package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_SecurityCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup CORS with empty origins (defaults to wildcard "*")
	// The current implementation sets AllowCredentials: true unconditionally
	router.Use(CORSMiddleware([]string{}))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Malicious site trying to access API
	origin := "https://evil.com"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check headers
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	allowCredentials := w.Header().Get("Access-Control-Allow-Credentials")

	t.Logf("Origin: %s", origin)
	t.Logf("Access-Control-Allow-Origin: %s", allowOrigin)
	t.Logf("Access-Control-Allow-Credentials: %s", allowCredentials)

	// If AllowCredentials is true, AllowOrigin MUST NOT be "*" (spec)
	// But if AllowOrigin reflects "https://evil.com" AND AllowCredentials is true,
	// then we have a security issue because we defaulted to "*" (allow all) but are allowing credentials.

	if allowCredentials == "true" {
		// If credentials allowed, we must ensure we are NOT allowing arbitrary origins
		// Since we initialized with empty list (which becomes *), if it works for evil.com, it's bad.
		if allowOrigin == "*" || allowOrigin == origin {
			assert.Fail(t, "Vulnerable configuration: AllowCredentials=true with Wildcard Origin (or reflected origin)")
		}
	}
}
