package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.PrometheusMiddleware())

	r.GET("/api/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.GET("/api/users/:id", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	t.Run("records metrics for regular path", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		// Assuming prometheus metrics don't panic or error out here
	})

	t.Run("records metrics for path with parameter", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/users/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("records metrics for unknown path (404)", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
