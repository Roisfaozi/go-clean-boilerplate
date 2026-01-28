package scenarios

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_AdvancedRateLimit_Tiers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	if !setup.IsDockerAvailable() {
		t.Skip("Skipping integration test: Docker not available")
	}

	// 1. Setup Environment (Redis is key here)
	redisContainer, redisPort, err := setup.SetupRedisContainer(context.Background())
	if err != nil {
		t.Skipf("Skipping integration test: Failed to start Redis container: %v", err)
		return
	}
	defer func() {
		_ = redisContainer.Terminate(context.Background())
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%s", redisPort),
	})
	require.NoError(t, redisClient.Ping(context.Background()).Err())

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Keep output clean

	// 2. Setup Router with middleware using the NEW tiered logic
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Define limiters exactly as in router.go
	publicLimiter := middleware.RateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeIP, 2, 60*time.Second) // 2 RPS for test
	criticalLimiter := middleware.RateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeIP, 1, 60*time.Second) // 1 RPM for test
	authLimiter := middleware.RateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeUser, 5, 60*time.Second) // 5 RPS for test

	// Mock Routes
	router.GET("/public", publicLimiter, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.POST("/login", criticalLimiter, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// For auth route, we need to mock setting the userID in context (simulating AuthMiddleware)
	router.GET("/dashboard", func(c *gin.Context) {
		// Mock Auth Middleware
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("userID", userID)
		}
		c.Next()
	}, authLimiter, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 3. Test Tier 1: Public API (IP Based)
	t.Run("Public_API_Limit", func(t *testing.T) {
		// Req 1: OK
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/public", nil)
		req.RemoteAddr = "1.1.1.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Req 2: OK
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/public", nil)
		req.RemoteAddr = "1.1.1.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Req 3: Limit Exceeded
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/public", nil)
		req.RemoteAddr = "1.1.1.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// Different IP should be OK
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/public", nil)
		req.RemoteAddr = "2.2.2.2:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 4. Test Tier 3: Critical Endpoint (Login - Strict IP)
	t.Run("Critical_Endpoint_Limit", func(t *testing.T) {
		// Req 1: OK
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "3.3.3.3:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Req 2: Limit Exceeded (Limit is 1)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "3.3.3.3:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	// 5. Test Tier 2: Authenticated User (User ID Based)
	t.Run("Authenticated_User_Limit", func(t *testing.T) {
		userID := "user-123"

		// Req 1-5: OK
		for i := 0; i < 5; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/dashboard", nil)
			req.Header.Set("X-User-ID", userID)
			req.RemoteAddr = "4.4.4.4:1234" // IP doesn't matter for user limit
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("Request %d should succeed", i+1))
		}

		// Req 6: Limit Exceeded
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("X-User-ID", userID)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// Different User should be OK (Independent Limits)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/dashboard", nil)
		req.Header.Set("X-User-ID", "user-456")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
