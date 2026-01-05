package http

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type AuthController struct {
	AuthUseCase usecase.AuthUseCase
	log         *logrus.Logger
	validate    *validator.Validate
}

func NewAuthController(useCase usecase.AuthUseCase, log *logrus.Logger, validate *validator.Validate) *AuthController {
	return &AuthController{
		AuthUseCase: useCase,
		log:         log,
		validate:    validate,
	}
}

// Login handles user login
func (h *AuthController) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithContext(c.Request.Context()).WithError(err).Error("Login failed: could not bind request")
		response.BadRequest(c, err, "could not bind request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.log.WithContext(c.Request.Context()).WithError(err).Error("Login failed: validation error")
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation error"), msg)
		return
	}

	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	res, refreshToken, err := h.AuthUseCase.Login(c.Request.Context(), req)
	if err != nil {
		h.log.WithContext(c.Request.Context()).Errorf("Login failed for user: %s", req.Username)
		response.HandleError(c, err, "Login failed")
		return
	}

	// Set refresh token in HttpOnly cookie
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", false, true)

	response.Success(c, res)
}

// RefreshToken handles token refresh
func (h *AuthController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		h.log.WithContext(c.Request.Context()).Warn("Refresh token not found in cookie")
		response.Unauthorized(c, exception.ErrUnauthorized, "refresh token not found")
		return
	}

	res, newRefreshToken, err := h.AuthUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		response.HandleError(c, err, "Refresh token failed")
		return
	}

	c.SetCookie("refresh_token", newRefreshToken, 3600*24*30, "/", "", false, true)
	response.Success(c, res)
}

// Logout handles user logout
func (h *AuthController) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")
	sessionID, _ := c.Get("sessionID")

	if userID == nil || sessionID == nil {
		response.Unauthorized(c, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	err := h.AuthUseCase.RevokeToken(c.Request.Context(), userID.(string), sessionID.(string))
	if err != nil {
		response.HandleError(c, err, "Logout failed")
		return
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	response.Success(c, gin.H{"message": "logged out successfully"})
}

// ForgotPassword handles forgot password request
func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.AuthUseCase.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		response.HandleError(c, err, "failed to process forgot password request")
		return
	}

	// Always return success for security reasons (don't reveal if email exists)
	response.Success(c, gin.H{"message": "If the email is registered, a reset link will be sent shortly."})
}

// ResetPassword handles password reset using token
func (h *AuthController) ResetPassword(c *gin.Context) {
	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.AuthUseCase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		response.HandleError(c, err, "failed to reset password")
		return
	}

	response.Success(c, gin.H{"message": "password reset successfully"})
}