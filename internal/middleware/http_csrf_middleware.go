package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

// HTTPCSRFMiddleware validates CSRF tokens on mutating requests when cookie auth is used.
func HTTPCSRFMiddleware(log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Authorization") != "" || r.Header.Get("X-API-Key") != "" {
				next.ServeHTTP(w, r)
				return
			}

			cookieToken, err := r.Cookie("csrf_token")
			headerToken := r.Header.Get("X-CSRF-Token")
			if err != nil || cookieToken.Value == "" || headerToken == "" ||
				subtle.ConstantTimeCompare([]byte(cookieToken.Value), []byte(headerToken)) != 1 {
				if log != nil {
					log.WithFields(logrus.Fields{
						"path":   r.URL.Path,
						"method": r.Method,
					}).Warn("CSRF token validation failed")
				}
				response.WriteError(w, http.StatusForbidden, errors.New("csrf token mismatch"), "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
