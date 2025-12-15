package http

import (
	"errors"
	"strings"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	AuthUseCase usecase.AuthUseCase
	Log         *logrus.Logger
	validate    *validator.Validate
}

func NewAuthHandler(authUseCase usecase.AuthUseCase, log *logrus.Logger, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		AuthUseCase: authUseCase,
		Log:         log,
		validate:    validate,
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
// @Failure      400      {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422      {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      401      {object}  response.SwaggerErrorResponseWrapper "Invalid credentials"
// @Failure      500      {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("Login failed: could not bind request")
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		h.Log.WithError(err).Error("Login failed: validation error")
		response.ValidationError(c, exception.ErrValidationError, msg)
		return
	}

	loginResp, refreshToken, err := h.AuthUseCase.Login(c.Request.Context(), req)
	if err != nil {
		h.Log.Errorf("Login failed for user: %s", req.Username)
		h.Log.WithError(err).Error("Login failed")
		h.handleError(c, err, "Wrong password or username")
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
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Refresh token not found or invalid"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		h.Log.Warn("Refresh token not found in cookie")
		response.Unauthorized(c, exception.ErrUnauthorized, "refresh token not found")
		return
	}

	tokenResp, newRefreshToken, err := h.AuthUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		h.Log.WithError(err).Error("Refresh token failed")
		h.handleError(c, err, "failed to refresh token")
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
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "User not authenticated or invalid session"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, exception.ErrUnauthorized, "unauthorized")
		return
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid session")
		return
	}

	err := h.AuthUseCase.RevokeToken(c.Request.Context(), userID.(string), sessionID.(string))
	if err != nil {
		h.handleError(c, err, "failed to revoke token")
		return
	}

	h.setRefreshTokenCookie(c, "")
	response.Success(c, gin.H{"message": "logged out successfully"})
}

func (h *AuthHandler) handleError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, usecase.ErrInvalidCredentials):
		response.Unauthorized(c, err, message)
	case errors.Is(err, usecase.ErrInvalidToken), errors.Is(err, usecase.ErrExpiredToken), errors.Is(err, usecase.ErrTokenRevoked):
		response.Unauthorized(c, err, message)
	case strings.Contains(err.Error(), "validation"):
		response.BadRequest(c, err, message)
	default:
		response.InternalServerError(c, err, message)
	}
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	var maxAge int
	if token == "" {
		maxAge = -1
	} else {
		maxAge = 3600 * 24 * 7
	}

	// Automatically set Secure flag in Release mode (Production)
	secure := gin.Mode() == gin.ReleaseMode
	c.SetCookie(
		"refresh_token",
		token,
		maxAge,
		"/api/v1/auth/refresh",
		"",
		secure,
		true,
	)
}
