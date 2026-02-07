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

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns access token and user info.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.LoginRequest true "Login request"
// @Success      200  {object}  response.SwaggerLoginResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/login [post]
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

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Refreshes access and refresh tokens using the refresh token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerTokenResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/refresh [post]
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

// Logout godoc
// @Summary      Logout user
// @Description  Revokes the current session and clears refresh token cookie.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/logout [post]
func (h *AuthController) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	sessionID, _ := c.Get("session_id")

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

// ForgotPassword godoc
// @Summary      Request password reset
// @Description  Sends a password reset email if the account exists.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.ForgotPasswordRequest true "Forgot password request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/forgot-password [post]
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

// ResetPassword godoc
// @Summary      Reset password
// @Description  Resets the user's password using a valid reset token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.ResetPasswordRequest true "Reset password request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/reset-password [post]
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

// VerifyEmail godoc
// @Summary      Verify email address
// @Description  Verifies the user's email address using a verification token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.VerifyEmailRequest true "Verify email request"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/verify-email [post]
func (h *AuthController) VerifyEmail(c *gin.Context) {
	var req model.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation failed"), msg)
		return
	}

	err := h.AuthUseCase.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		response.HandleError(c, err, "failed to verify email")
		return
	}

	response.Success(c, gin.H{"message": "email verified successfully"})
}

// ResendVerification godoc
// @Summary      Resend verification email
// @Description  Resends the email verification link to the authenticated user.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Already verified or request failed"
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/resend-verification [post]
func (h *AuthController) ResendVerification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists || userID == nil {
		response.Unauthorized(c, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	err := h.AuthUseCase.RequestVerification(c.Request.Context(), userID.(string))
	if err != nil {
		response.HandleError(c, err, "failed to request verification email")
		return
	}

	response.Success(c, gin.H{"message": "verification email sent successfully"})
}

// Register godoc
// @Summary      Register new user
// @Description  Creates a new user account and auto-provisions a default workspace.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterRequest true "Registration request"
// @Success      201  {object}  response.SwaggerLoginResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper "Invalid request body"
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper "Validation Error"
// @Failure      409  {object}  response.SwaggerErrorResponseWrapper "Username or Email already exists"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/register [post]
func (h *AuthController) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithContext(c.Request.Context()).WithError(err).Error("Register failed: could not bind request")
		response.BadRequest(c, err, "could not bind request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.log.WithContext(c.Request.Context()).WithError(err).Error("Register failed: validation error")
		msg := validation.FormatValidationErrors(err)
		response.ValidationError(c, errors.New("validation error"), msg)
		return
	}

	req.IPAddress = c.ClientIP()
	req.UserAgent = c.Request.UserAgent()

	res, refreshToken, err := h.AuthUseCase.Register(c.Request.Context(), req)
	if err != nil {
		h.log.WithContext(c.Request.Context()).Errorf("Register failed for user: %s", req.Username)
		response.HandleError(c, err, "Register failed")
		return
	}

	// Set refresh token in HttpOnly cookie
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", false, true)

	response.Created(c, res)
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the currently authenticated user's information.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Router       /auth/me [get]
func (h *AuthController) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("user_role")

	if userID == nil {
		response.Unauthorized(c, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	response.Success(c, gin.H{
		"user": gin.H{
			"id":       userID,
			"username": username,
			"role":     role,
		},
	})
}
