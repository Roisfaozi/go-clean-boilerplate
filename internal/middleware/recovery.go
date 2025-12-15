package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RecoveryMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())

				requestID := c.GetString("request_id")
				if requestID == "" {
					requestID = c.GetHeader("X-Request-ID")
				}

				log.WithFields(logrus.Fields{
					"type":        "panic_recovery",
					"request_id":  requestID,
					"error":       err,
					"stack_trace": stack,
					"path":        c.Request.URL.Path,
					"method":      c.Request.Method,
				}).Error("Panic recovered")

				// In production, we don't want to expose the detailed error.
				// For now, we just pass a generic error message.
				// The actual error is logged above.
				response.InternalServerError(c, fmt.Errorf("internal server error"), "Something went wrong")
				c.Abort()
			}
		}()
		c.Next()
	}
}
