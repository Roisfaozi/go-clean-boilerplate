package middleware

import (
	"errors"
	"strings"

	authUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/authcontext"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	headerAuthorization = "Authorization"
	authSchemeBearer    = "bearer"
	cookieAccessToken   = "access_token"
)

type AuthMiddleware struct {
	AuthUseCase   authUsecase.AuthUseCase
	Log           *logrus.Logger
	TicketManager ws.TicketManager
}

func NewAuthMiddleware(authUseCase authUsecase.AuthUseCase, log *logrus.Logger, ticketManager ws.TicketManager) *AuthMiddleware {
	return &AuthMiddleware{
		AuthUseCase:   authUseCase,
		Log:           log,
		TicketManager: ticketManager,
	}
}

func (m *AuthMiddleware) ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAPIKeyAuth(c) {
			c.Next()
			return
		}

		token := ""
		authHeader := c.GetHeader(headerAuthorization)
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == authSchemeBearer {
				token = parts[1]
			}
		}

		// Fallback: Check for access_token cookie
		if token == "" {
			cookieToken, err := c.Cookie(cookieAccessToken)
			if err == nil && cookieToken != "" {
				token = cookieToken
			}
		}

		if token == "" {
			response.Unauthorized(c, errors.New("token is required"), "unauthorized")
			c.Abort()
			return
		}

		claims, err := m.AuthUseCase.ValidateAccessToken(token)
		if err != nil {
			m.Log.WithError(err).Warn("Token validation failed")
			response.Unauthorized(c, err, "unauthorized")
			c.Abort()
			return
		}

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

		c.Set(contextKeyUserID, claims.UserID)
		c.Set(contextKeySessionID, claims.SessionID)
		c.Set(contextKeyUserRole, claims.Role)
		c.Set(contextKeyUsername, claims.Username)
		c.Set(authMethodContextKey, authMethodJWT)

		ctx := authcontext.WithUserID(c.Request.Context(), claims.UserID)
		ctx = authcontext.WithSessionID(ctx, claims.SessionID)
		ctx = authcontext.WithRole(ctx, claims.Role)
		ctx = authcontext.WithUsername(ctx, claims.Username)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (m *AuthMiddleware) ValidateWebSocketToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ticket := c.Query("ticket")
		if ticket == "" {
			response.Unauthorized(c, errors.New("ticket is required"), "unauthorized")
			c.Abort()
			return
		}

		userCtx, err := m.TicketManager.ValidateTicket(c.Request.Context(), ticket)
		if err != nil {
			m.Log.WithError(err).Warn("Invalid or expired WebSocket ticket")
			response.Unauthorized(c, errors.New("invalid or expired ticket"), "unauthorized")
			c.Abort()
			return
		}

		c.Set(contextKeyUserID, userCtx.UserID)
		c.Set(contextKeySessionID, userCtx.SessionID)
		c.Set(contextKeyUserRole, userCtx.Role)
		c.Set(contextKeyUsername, userCtx.Username)

		ctx := authcontext.WithUserID(c.Request.Context(), userCtx.UserID)
		ctx = authcontext.WithSessionID(ctx, userCtx.SessionID)
		ctx = authcontext.WithRole(ctx, userCtx.Role)
		ctx = authcontext.WithUsername(ctx, userCtx.Username)

		// Context from ticket takes precedence.
		if userCtx.OrganizationID != "" {
			c.Set(contextKeyOrganizationID, userCtx.OrganizationID)
		}
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

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
