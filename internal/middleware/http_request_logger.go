package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/sirupsen/logrus"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	size        int
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.statusCode = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.statusCode = http.StatusOK
		r.wroteHeader = true
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseRecorder) Flush() {
	if fl, ok := r.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return hj.Hijack()
}

// hijackForwarder preserves the inner ResponseWriter's optional interfaces
// (Hijacker, Flusher, Pusher, CloseNotifier) so WebSocket upgrades keep working.
type hijackForwarder struct {
	http.ResponseWriter
}

func (f hijackForwarder) Flush() {
	if fl, ok := f.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (f hijackForwarder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := f.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return hj.Hijack()
}

func (f hijackForwarder) Unwrap() http.ResponseWriter {
	return f.ResponseWriter
}

// HTTPRequestLogger logs inbound requests and outcoming response statuses.
func HTTPRequestLogger(log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			rec := &responseRecorder{ResponseWriter: hijackForwarder{w}, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			latency := time.Since(startTime)
			if log != nil {
				route := r.Pattern
				if route == "" {
					route = "unknown"
				}
				authMethod, _ := delivery.GetContextString(r.Context(), delivery.AuthMethodKey)
				clientIP := extractClientIP(r)
				if clientIP == "" {
					clientIP = r.RemoteAddr
				}

				entry := log.WithContext(r.Context()).WithFields(logrus.Fields{
					"type":           "http_request",
					"method":         r.Method,
					"path":           r.URL.Path,
					"route":          route,
					"auth_method":    authMethod,
					"status":         rec.statusCode,
					"latency_ms":     float64(latency.Microseconds()) / 1000.0,
					"client_ip":      clientIP,
					"user_agent":     r.UserAgent(),
					"data_length":    rec.size,
					"response_bytes": rec.size,
				})

				switch {
				case rec.statusCode >= http.StatusInternalServerError:
					entry.Error("Internal Server Error")
				case rec.statusCode >= http.StatusBadRequest:
					entry.Warn("Client Error")
				default:
					entry.Info("Request Processed")
				}
			}
		})
	}
}
