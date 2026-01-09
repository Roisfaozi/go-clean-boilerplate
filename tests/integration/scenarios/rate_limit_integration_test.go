//go:build integration
// +build integration

package scenarios

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestScenario_RateLimit_Redis_Distributed verifies that the Redis-based rate limiter
// correctly blocks requests exceeding the defined limit.
func TestScenario_RateLimit_Redis_Distributed(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	// No need to cleanup DB as this relies on Redis only, but good practice
	setup.CleanupDatabase(t, env.DB)

	// 1. Setup Middleware with LOW limit
	// RPS = 0.1 => Limit = 0.1 * 60 = 6 requests per minute
	// Let's use a slightly higher RPS to allow a few requests, then block.
	// 5 RPS => 300/min. Too high.
	// We want to block after say 3 requests.
	// The middleware logic: limit = int64(rps * 60).
	// If we want limit = 3, then rps = 3/60 = 0.05.
	// But limit < 1 becomes 1.

	// Let's use RPS = 0.05 (Limit = 3 requests per minute)
	rps := 0.05
	expectedLimit := int64(3) // 0.05 * 60 = 3

	// 2. Setup Router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RateLimitMiddlewareRedis(env.Redis, env.Logger, rps))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 3. Simulate Requests
	// First 3 requests should pass
	for i := 0; i < int(expectedLimit); i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		// Important: ClientIP is determined by remote addr or headers.
		// Default httptest doesn't set headers, assumes local.
		// Since we use the same "client", IP stays consistent.
		req.RemoteAddr = "192.168.1.100:1234"

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass", i+1)
	}

	// 4. Verify 4th Request is BLOCKED
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:1234" // Same IP
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request exceeding limit should be blocked")

	// 5. Verify DIFFERENT IP is NOT blocked
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.5:5678" // Different IP
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Request from different IP should pass")
}
