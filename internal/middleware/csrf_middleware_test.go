package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	t.Run("Safe method GET allowed without tokens", func(t *testing.T) {
		r := gin.New()
		r.Use(CSRFMiddleware(logger))
		r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Authorization header bypasses CSRF", func(t *testing.T) {
		r := gin.New()
		r.Use(CSRFMiddleware(logger))
		r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.Header.Set("Authorization", "Bearer xyz")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Cookie without matching header returns forbidden", func(t *testing.T) {
		r := gin.New()
		r.Use(CSRFMiddleware(logger))
		r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-123"})
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Matching cookie and header succeeds", func(t *testing.T) {
		r := gin.New()
		r.Use(CSRFMiddleware(logger))
		r.POST("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-123"})
		req.Header.Set("X-CSRF-Token", "token-123")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
