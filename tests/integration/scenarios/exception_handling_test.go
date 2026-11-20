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

func TestScenario_ExceptionHandling_PanicRecovery(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)

	router := http.NewServeMux()
	router.Handle("GET /panic-trigger", delivery.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("intentional crash for testing")
		}),
		middleware.HTTPRequestIDMiddleware(),
		middleware.HTTPRecoveryMiddleware(env.Logger),
	))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic-trigger", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected 500 Internal Server Error after panic")

}
