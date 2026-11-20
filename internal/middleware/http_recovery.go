package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"runtime/debug"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

type recoveryResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *recoveryResponseWriter) WriteHeader(code int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *recoveryResponseWriter) Write(b []byte) (int, error) {
	w.wroteHeader = true
	return w.ResponseWriter.Write(b)
}

func (w *recoveryResponseWriter) Flush() {
	if fl, ok := w.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (w *recoveryResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return hj.Hijack()
}

func (w *recoveryResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// HTTPRecoveryMiddleware catches and recovers panics during request execution.
func HTTPRecoveryMiddleware(log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &recoveryResponseWriter{ResponseWriter: w}
			defer func() {
				if err := recover(); err != nil {
					stack := string(debug.Stack())
					reqID, _ := r.Context().Value(constants.RequestIDKey).(string)

					if log != nil {
						log.WithContext(r.Context()).WithFields(logrus.Fields{
							"type":        "panic_recovery",
							"request_id":  reqID,
							"error":       err,
							"stack_trace": stack,
							"path":        r.URL.Path,
							"method":      r.Method,
						}).Error("Panic recovered")
					}

					if !rw.wroteHeader {
						response.WriteError(w, http.StatusInternalServerError, errors.New("internal server error"), "Something went wrong")
					}
				}
			}()
			next.ServeHTTP(rw, r)
		})
	}
}
