package middleware

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			uid, _ := uuid.NewV7()
			requestID = uid.String()
		}

		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Set(string(constants.RequestIDKey), requestID)

		ctx := context.WithValue(c.Request.Context(), constants.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
