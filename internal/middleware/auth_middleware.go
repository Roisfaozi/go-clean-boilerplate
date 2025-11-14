package middleware

import (
	"errors"
	"strings"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	AuthUseCase usecase.AuthUseCase
	Log         *logrus.Logger
}

func NewAuthMiddleware(authUseCase usecase.AuthUseCase, log *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		AuthUseCase: authUseCase,
		Log:         log,
	}
}

// ValidateToken validates the JWT token from the Authorization header
func (m *AuthMiddleware) ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, errors.New("authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, errors.New("invalid authorization header format"))
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, errors.New("token is required"))
			c.Abort()
			return
		}

		claims, err := m.AuthUseCase.ValidateAccessToken(token)
		if err != nil {
			m.Log.WithError(err).Warn("Token validation failed")
			// We can pass the specific error from the use case directly
			response.Unauthorized(c, err)
			c.Abort()
			return
		}

		// Verify session is still valid in the repository
		session, err := m.AuthUseCase.Verify(c.Request.Context(), claims.UserID, claims.SessionID)
		if err != nil {
			m.Log.WithError(err).Warn("Session verification failed with database/redis error")
			response.InternalServerError(c, errors.New("could not verify session"))
			c.Abort()
			return
		}
		if session == nil {
			m.Log.Warn("Session is not valid or has been revoked")
			response.Unauthorized(c, errors.New("invalid or expired session"))
			c.Abort()
			return
		}

		// Set user and session info in the context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("session_id", claims.SessionID)

		c.Next()
	}
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		return "", false
	}

	return userIDStr, true
}

// GetSessionIDFromContext retrieves the session ID from the context
func GetSessionIDFromContext(c *gin.Context) (string, bool) {
	sessionID, exists := c.Get("session_id")
	if !exists {
		return "", false
	}

	sessionIDStr, ok := sessionID.(string)
	if !ok || sessionIDStr == "" {
		return "", false
	}

	return sessionIDStr, true
}
