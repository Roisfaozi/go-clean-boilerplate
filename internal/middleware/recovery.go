package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RecoveryMiddleware captures panics and logs them structuredly
func RecoveryMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Capture stack trace
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

				// Return 500 to client
				response.InternalServerError(c, fmt.Errorf("internal server error"), "Something went wrong")
				c.Abort()
			}
		}()
		c.Next()
	}
}
