package http

import (
	"errors"
	"strings"

	"github.com/Roisfaozi/casbin-db/internal/config"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	AuthUseCase usecase.AuthUseCase
	Log         *logrus.Logger
	Config      *config.AppConfig
}

func NewAuthHandler(authUseCase usecase.AuthUseCase, log *logrus.Logger, cfg *config.AppConfig) *AuthHandler {
	return &AuthHandler{
		AuthUseCase: authUseCase,
		Log:         log,
		Config:      cfg,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Warn("Login failed: could not bind request")
		response.BadRequest(c, errors.New("invalid request body"))
		return
	}

	// The use case now handles validation, so we can remove it from here.
	loginResp, refreshToken, err := h.AuthUseCase.Login(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, refreshToken)
	response.Success(c, loginResp)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		h.Log.Warn("Refresh token not found in cookie")
		response.Unauthorized(c, errors.New("refresh token not found"))
		return
	}

	tokenResp, newRefreshToken, err := h.AuthUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.setRefreshTokenCookie(c, newRefreshToken)
	response.Success(c, tokenResp)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, errors.New("user not authenticated"))
		return
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		response.Unauthorized(c, errors.New("invalid session"))
		return
	}

	err := h.AuthUseCase.RevokeToken(c.Request.Context(), userID.(string), sessionID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Clear the refresh token cookie
	h.setRefreshTokenCookie(c, "")
	response.Success(c, gin.H{"message": "logged out successfully"})
}

// handleError centralizes error handling for the auth handler
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	h.Log.WithError(err).Error("An error occurred in auth handler")
	switch {
	case errors.Is(err, usecase.ErrInvalidCredentials):
		response.Unauthorized(c, err)
	case errors.Is(err, usecase.ErrInvalidToken), errors.Is(err, usecase.ErrExpiredToken), errors.Is(err, usecase.ErrTokenRevoked):
		response.Unauthorized(c, err)
	// Catch validation errors (this requires your validator to be configured to return error)
	case strings.Contains(err.Error(), "validation"):
		response.BadRequest(c, err)
	default:
		response.InternalServerError(c, errors.New("an unexpected internal error occurred"))
	}
}

// setRefreshTokenCookie sets or clears the refresh token in an HTTP-only cookie
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	var maxAge int
	if token == "" {
		maxAge = -1 // Expire immediately
	} else {
		maxAge = 3600 * 24 * 7 // 7 days
	}

	secure := true

	if h.Config.Server.AppEnv != "production" {
		maxAge = 3600 * 24 * 7 // 7 days
		secure = false
		c.SetCookie(
			"refresh_token",
			token,
			maxAge,
			"/api/v1/auth/refresh", // Path should be specific to the refresh endpoint
			"",                     // Domain
			secure,                 // Secure flag (true in production)
			true,                   // HttpOnly flag
		)
		return
	}
	c.SetCookie(
		"refresh_token",
		token,
		maxAge,
		"/api/v1/auth/refresh", // Path should be specific to the refresh endpoint
		"",                     // Domain
		secure,                 // Secure flag (true in production)
		true,                   // HttpOnly flag
	)
}
