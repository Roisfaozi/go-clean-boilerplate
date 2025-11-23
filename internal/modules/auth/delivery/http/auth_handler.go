package http

import (
	"errors"
	"strings"

	"github.com/Roisfaozi/casbin-db/internal/modules/auth/model"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	AuthUseCase usecase.AuthUseCase
	Log         *logrus.Logger
}

func NewAuthHandler(authUseCase usecase.AuthUseCase, log *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		AuthUseCase: authUseCase,
		Log:         log,
	}
}

// Login handles user login
// @Summary      User Login
// @Description  Logs in a user by validating credentials and returns an access token and a refresh token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      model.LoginRequest  true  "Login Credentials"
// @Success      200      {object}  response.SwaggerLoginResponseWrapper
// @Failure      400      {object}  response.WebResponseAny "Invalid request body"
// @Failure      401      {object}  response.WebResponseAny "Invalid credentials"
// @Failure      500      {object}  response.WebResponseAny "Internal server error"
// @Router       /auth/login [post]
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
// @Summary      Refresh Access Token
// @Description  Refreshes an access token using a valid refresh token provided in an HTTP-only cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  response.SwaggerTokenResponseWrapper
// @Failure      401  {object}  response.WebResponseAny "Refresh token not found or invalid"
// @Failure      500  {object}  response.WebResponseAny "Internal server error"
// @Router       /auth/refresh [post]
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
// @Summary      User Logout
// @Description  Logs out the current user by revoking their session.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "message: logged out successfully"
// @Failure      401  {object}  response.WebResponseAny "User not authenticated or invalid session"
// @Failure      500  {object}  response.WebResponseAny "Internal server error"
// @Router       /auth/logout [post]
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

	secure := false

	//if h.Config.Server.AppEnv != "production" {
	//	maxAge = 3600 * 24 * 7 // 7 days
	//	secure = false
	//	c.SetCookie(
	//		"refresh_token",
	//		token,
	//		maxAge,
	//		"/api/v1/auth/refresh", // Path should be specific to the refresh endpoint
	//		"",                     // Domain
	//		secure,                 // Secure flag (true in production)
	//		true,                   // HttpOnly flag
	//	)
	//	return
	//}
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