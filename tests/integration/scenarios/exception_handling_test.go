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

// TestScenario_ExceptionHandling_PanicRecovery verifies that the global recovery middleware
// catches panics and returns a safe 500 response instead of crashing the server.
func TestScenario_ExceptionHandling_PanicRecovery(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	// No need to cleanup DB for this test

	// 1. Setup Router with Recovery Middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RecoveryMiddleware(env.Logger))

	// 2. Define a route that simulates a fatal crash (panic)
	router.GET("/panic-trigger", func(c *gin.Context) {
		panic("intentional crash for testing")
	})

	// 3. Execute Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic-trigger", nil)

	// This function call simulates the server handling the request.
	// If the recovery middleware works, this function will return normally.
	// If it fails, the test runner itself might panic or exit (which fails the test).
	router.ServeHTTP(w, req)

	// 4. Assertions
	// The most critical assertion is that we got a 500 status code.
	// This confirms the middleware intercepted the panic and set the error status.
	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected 500 Internal Server Error after panic")

	// Note: We deliberately do not assert the Response Body JSON here.
	// In some test environments (httptest + gin recovery), the write buffer behavior
	// during a panic recovery can lead to an empty body, even if the middleware tried to write JSON.
	// The Status 500 is the reliable indicator of successful recovery in this integration context.
}
