package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware_Comprehensive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Helper to create router with middleware
	setupRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(SecurityMiddleware())
		return r
	}

	t.Run("Positive Case - Standard GET Request", func(t *testing.T) {
		r := setupRouter()
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'self'; frame-ancestors 'none';", w.Header().Get("Content-Security-Policy"))
		assert.Equal(t, "geolocation=(), microphone=(), camera=(), payment=()", w.Header().Get("Permissions-Policy"))
	})

	t.Run("Edge Case - Different HTTP Methods (POST)", func(t *testing.T) {
		r := setupRouter()
		r.POST("/submit", func(c *gin.Context) {
			c.Status(http.StatusCreated)
		})

		req, _ := http.NewRequest("POST", "/submit", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		// Headers should still be present regardless of method or status code
		assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
		assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
	})

	t.Run("Negative Case - Middleware Not Applied", func(t *testing.T) {
		// Control group: Verify headers are NOT present if middleware is missing
		r := gin.New()
		r.GET("/insecure", func(c *gin.Context) {
			c.String(http.StatusOK, "Unsafe")
		})

		req, _ := http.NewRequest("GET", "/insecure", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Security-Policy"), "CSP should be missing without middleware")
		assert.Empty(t, w.Header().Get("X-Frame-Options"), "X-Frame-Options should be missing without middleware")
		assert.Empty(t, w.Header().Get("Permissions-Policy"), "Permissions-Policy should be missing without middleware")
	})

	t.Run("Vulnerability Case - Prevent Weak CSP Configurations", func(t *testing.T) {
		r := setupRouter()
		r.GET("/vuln-check", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("GET", "/vuln-check", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		csp := w.Header().Get("Content-Security-Policy")

		// Vulnerability Check 1: Ensure we are NOT allowing 'unsafe-inline' (XSS Risk)
		if strings.Contains(csp, "'unsafe-inline'") {
			t.Errorf("Security Vulnerability: CSP allows 'unsafe-inline'")
		}

		// Vulnerability Check 2: Ensure we are NOT allowing 'unsafe-eval' (XSS Risk)
		if strings.Contains(csp, "'unsafe-eval'") {
			t.Errorf("Security Vulnerability: CSP allows 'unsafe-eval'")
		}

		// Vulnerability Check 3: Ensure frame-ancestors is restricted (Clickjacking Risk)
		// We expect 'none' or specific origins, but definitely not '*' or missing
		if !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("Security Vulnerability: CSP does not restrict frame ancestors correctly (Clickjacking risk). Got: %s", csp)
		}
	})
}
