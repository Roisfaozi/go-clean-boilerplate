package middleware

import (
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
)

// HTTPSecurityMiddleware adds defensive security headers to all HTTP responses.
func HTTPSecurityMiddleware() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none';")
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

			next.ServeHTTP(w, r)
		})
	}
}
