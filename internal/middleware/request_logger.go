package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequestLogger middleware handles structured logging for HTTP requests.
// It relies on RequestIDMiddleware being called before it to inject the Trace ID.
func RequestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		userAgent := c.Request.UserAgent()
		dataLength := c.Writer.Size()

		if dataLength < 0 {
			dataLength = 0
		}

		// Use WithContext(c.Request.Context()) to automatically pick up the request_id via TraceContextHook
		entry := log.WithContext(c.Request.Context()).WithFields(logrus.Fields{
			"type":        "http_request",
			"method":      method,
			"path":        path,
			"status":      statusCode,
			"latency_ns":  latency.Nanoseconds(),
			"latency_ms":  float64(latency.Nanoseconds()) / 1e6,
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"data_length": dataLength,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			if statusCode >= 500 {
				entry.Error("Internal Server Error")
			} else if statusCode >= 400 {
				entry.Warn("Client Error")
			} else {
				entry.Info("Request Processed")
			}
		}
	}
}
