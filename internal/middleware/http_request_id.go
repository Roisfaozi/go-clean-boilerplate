package middleware

import (
	"context"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/google/uuid"
)

// HTTPRequestIDMiddleware injects X-Request-ID into request context and response header.
func HTTPRequestIDMiddleware() delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(headerXRequestID)
			if requestID == "" {
				uid, _ := uuid.NewV7()
				requestID = uid.String()
			}

			w.Header().Set(headerXRequestID, requestID)
			ctx := context.WithValue(r.Context(), constants.RequestIDKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
