//go:build integration
// +build integration

package scenarios

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
)

func TestScenario_RateLimit_Redis_Distributed(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	rps := 3
	windowSeconds := 60

	router := http.NewServeMux()
	router.Handle("GET /test", delivery.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		middleware.HTTPRateLimitMiddlewareRedis(env.Redis, env.Logger, middleware.LimiterTypeIP, rps, windowSeconds),
	))

	for i := 0; i < rps; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)

		req.RemoteAddr = "192.168.1.100:1234"

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass", i+1)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:1234"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request exceeding limit should be blocked")

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.5:5678"
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code, "Request from different IP should pass")
}
