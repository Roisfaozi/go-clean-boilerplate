package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CSRFMiddleware(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		if c.GetHeader("Authorization") != "" || c.GetHeader("X-API-Key") != "" {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie("csrf_token")
		headerToken := c.GetHeader("X-CSRF-Token")
		if err != nil || cookieToken == "" || headerToken == "" ||
			subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
			if log != nil {
				log.WithFields(logrus.Fields{
					"path":   c.Request.URL.Path,
					"method": c.Request.Method,
				}).Warn("CSRF token validation failed")
			}
			response.Forbidden(c, errors.New("csrf token mismatch"), "forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
