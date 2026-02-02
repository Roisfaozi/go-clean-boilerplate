package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Generates New ID", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestIDMiddleware())
		r.GET("/test-id", func(c *gin.Context) {
			id := c.Writer.Header().Get("X-Request-ID")
			assert.NotEmpty(t, id)

			// Check gin context
			ctxID, exists := c.Get(string(constants.RequestIDKey))
			assert.True(t, exists)
			assert.Equal(t, id, ctxID)

			// Check request context
			reqCtxID := c.Request.Context().Value(constants.RequestIDKey)
			assert.Equal(t, id, reqCtxID)

			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-id", nil)
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.NotEmpty(t, resp.Header().Get("X-Request-ID"))
	})

	t.Run("Uses Existing ID", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestIDMiddleware())
		r.GET("/test-id-exist", func(c *gin.Context) {
			id := c.Writer.Header().Get("X-Request-ID")
			assert.Equal(t, "existing-id-123", id)
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test-id-exist", nil)
		req.Header.Set("X-Request-ID", "existing-id-123")
		resp := httptest.NewRecorder()

		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "existing-id-123", resp.Header().Get("X-Request-ID"))
	})
}
