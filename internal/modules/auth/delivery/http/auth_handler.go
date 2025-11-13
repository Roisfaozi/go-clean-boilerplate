package http

import (
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
// @Summary User login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("Failed to bind request body")
		response.BadRequest(c, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.Log.WithError(err).Error("Invalid login request")
		response.BadRequest(c, err)
		return
	}

	loginResp, refreshToken, err := h.AuthUseCase.Login(ctx, req)
	if err != nil {
		h.handleError(c, err, "Login failed")
		return
	}

	// Set refresh token as HTTP-only cookie
	h.setRefreshTokenCookie(c, refreshToken, int(loginResp.ExpiresIn))

	response.Success(c, loginResp)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.RefreshRequest true "Refresh token"
// @Success 200 {object} model.TokenResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Log.WithError(err).Error("Failed to bind request body")
		response.BadRequest(c, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.Log.WithError(err).Error("Invalid refresh request")
		response.BadRequest(c, err)
		return
	}

	tokenResp, newRefreshToken, err := h.AuthUseCase.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.handleError(c, err, "Failed to refresh token")
		return
	}

	// Update refresh token cookie
	h.setRefreshTokenCookie(c, newRefreshToken, int(tokenResp.ExpiresIn))

	response.Success(c, tokenResp)
}

// Logout handles user logout
// @Summary User logout
// @Description Invalidate user session
// @Tags auth
// @Security Bearer
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		response.Unauthorized(c, "Invalid session")
		return
	}

	ctx := c.Request.Context()
	err := h.AuthUseCase.RevokeToken(ctx, userID.(string), sessionID.(string))
	if err != nil {
		h.handleError(c, err, "Failed to logout")
		return
	}

	// Clear refresh token cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1, // Expire immediately
		"/",
		"",
		false, // Secure
		true,  // HttpOnly
	)

	response.Success(c, nil)
}

// handleError handles common errors
func (h *AuthHandler) handleError(c *gin.Context, err error, message string) {
	h.Log.WithError(err).Error(message)

	switch {
	case err == usecase.ErrInvalidCredentials:
		response.Unauthorized(c, "Invalid credentials")
	case err == usecase.ErrInvalidToken || err == usecase.ErrExpiredToken:
		response.Unauthorized(c, err.Error())
	default:
		response.InternalServerError(c, err)
	}
}

// setRefreshTokenCookie sets the refresh token as an HTTP-only cookie
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	c.SetCookie(
		"refresh_token",
		token,
		maxAge,
		"/api/v1/auth/refresh",
		"",
		false, // Set to true in production with HTTPS
		true,  // HttpOnly
	)
}
