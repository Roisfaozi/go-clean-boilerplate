package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

// HTTPRecoveryMiddleware catches and recovers panics during request execution.
func HTTPRecoveryMiddleware(log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := string(debug.Stack())
					reqID, _ := r.Context().Value(constants.RequestIDKey).(string)

					if log != nil {
						log.WithFields(logrus.Fields{
							"type":        "panic_recovery",
							"request_id":  reqID,
							"error":       err,
							"stack_trace": stack,
							"path":        r.URL.Path,
							"method":      r.Method,
						}).Error("Panic recovered")
					}

					response.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"), "Something went wrong")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
