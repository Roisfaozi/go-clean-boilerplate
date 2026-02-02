package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Records Metrics", func(t *testing.T) {
		r := gin.New()
		r.Use(PrometheusMiddleware())
		r.GET("/test-metrics", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-metrics", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify metrics objects are initialized
		assert.NotNil(t, httpRequestsTotal)
		assert.NotNil(t, httpRequestDuration)
	})

	t.Run("Unknown Path", func(t *testing.T) {
		r := gin.New()
		r.Use(PrometheusMiddleware())

		// No route defined

		req, _ := http.NewRequest(http.MethodGet, "/unknown", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}
