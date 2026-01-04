package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		validate func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Headers are set correctly",
			validate: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
				assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
				assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
				assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
				assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/", nil)

			middleware.SecurityMiddleware()(c)

			tt.validate(t, w)
		})
	}
}
