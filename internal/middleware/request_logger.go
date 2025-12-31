package middleware

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestLogger middleware handles structured logging for HTTP requests
func RequestLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)

		c.Set("request_id", requestID)

		ctx := context.WithValue(c.Request.Context(), constants.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

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

		entry := log.WithFields(logrus.Fields{
			"type":        "http_request",
			"request_id":  requestID,
			"method":      method,
			"path":        path,
			"status":      statusCode,
			"latency_ns":  latency.Nanoseconds(),
			"latency_ms":  float64(latency.Nanoseconds()) / 1e6,
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"data_length": dataLength,
		})

		if userID, exists := c.Get("user_id"); exists {
			entry = entry.WithField("user_id", userID)
		}

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
