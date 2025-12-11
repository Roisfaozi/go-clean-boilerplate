package middleware

import (
	"errors"
	"strings"

	authUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase" // Alias to avoid conflict with jwt.Claims
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	AuthUseCase authUsecase.AuthUseCase
	Log         *logrus.Logger
}

func NewAuthMiddleware(authUseCase authUsecase.AuthUseCase, log *logrus.Logger) *AuthMiddleware {
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
			response.Unauthorized(c, errors.New("authorization header is required"), "unauthorized")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, errors.New("invalid authorization header format"), "unauthorized")
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			response.Unauthorized(c, errors.New("token is required"), "unauthorized")
			c.Abort()
			return
		}

		claims, err := m.AuthUseCase.ValidateAccessToken(token) // This returns *jwt.Claims from internal/pkg/jwt
		if err != nil {
			m.Log.WithError(err).Warn("Token validation failed")
			response.Unauthorized(c, err, "unauthorized")
			c.Abort()
			return
		}

		// Verify session is still valid in the repository
		session, err := m.AuthUseCase.Verify(c.Request.Context(), claims.UserID, claims.SessionID)
		if err != nil {
			m.Log.WithError(err).Warn("Session verification failed with database/redis error")
			response.InternalServerError(c, errors.New("could not verify session"), "internal server error")
			c.Abort()
			return
		}
		if session == nil {
			m.Log.Warn("Session is not valid or has been revoked")
			response.Unauthorized(c, errors.New("invalid or expired session"), "unauthorized")
			c.Abort()
			return
		}

		// Set user, session, role, and username info in the context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("session_id", claims.SessionID)
		c.Set("user_role", claims.Role)
		c.Set("username", claims.Username)

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

// GetRoleFromContext retrieves the user role from the context
func GetRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	if !ok || roleStr == "" {
		return "", false
	}
	return roleStr, true
}

// GetUsernameFromContext retrieves the username from the context
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	usernameStr, ok := username.(string)
	if !ok || usernameStr == "" {
		return "", false
	}
	return usernameStr, true
}
