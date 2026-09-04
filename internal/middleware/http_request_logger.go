package middleware

import (
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/sirupsen/logrus"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// HTTPRequestLogger logs inbound requests and outcoming response statuses.
func HTTPRequestLogger(log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			latency := time.Since(startTime)
			if log != nil {
				log.WithContext(r.Context()).WithFields(logrus.Fields{
					"type":        "http_request",
					"method":      r.Method,
					"path":        r.URL.Path,
					"status":      rec.statusCode,
					"latency_ms":  float64(latency.Microseconds()) / 1000.0,
					"client_ip":   r.RemoteAddr,
					"user_agent":  r.UserAgent(),
					"data_length": rec.size,
				}).Info("Request Processed")
			}
		})
	}
}
