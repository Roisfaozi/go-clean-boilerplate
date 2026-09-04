package middleware

import (
	"net/http"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
)

// HTTPCORSMiddleware handles CORS headers and preflight requests for net/http.
func HTTPCORSMiddleware(allowedOrigins []string) delivery.Middleware {
	allowAll := false
	originMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		if o == "*" {
			allowAll = true
		}
		originMap[o] = true
	}

	allowHeaders := strings.Join([]string{
		"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With",
		"X-Organization-ID", "X-Organization-Slug", "X-CSRF-Token", "X-API-Key",
		"Tus-Resumable", "Upload-Length", "Upload-Metadata", "Upload-Offset",
		"Upload-Protocol", "Upload-Draft-Interop-Version",
	}, ", ")

	exposeHeaders := strings.Join([]string{
		"Content-Length", "Upload-Offset", "Location", "Upload-Length",
		"Tus-Version", "Tus-Resumable", "Tus-Max-Size", "Tus-Extension", "Upload-Metadata",
	}, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if originMap[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
				w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
