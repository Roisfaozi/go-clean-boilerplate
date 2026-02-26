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
	// Set access token in HttpOnly cookie (short lived)
	c.SetCookie("access_token", res.AccessToken, int(res.ExpiresIn), "/", "", false, true)

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
	c.SetCookie("access_token", res.AccessToken, int(res.ExpiresIn), "/", "", false, true)
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
	userID, exists := c.Get("user_id")
	if !exists || userID == nil {
		response.Unauthorized(c, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid user context")
		return
	}

	sessionID, ok := GetSessionIDFromContext(c)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid session context")
		return
	}

	err := h.AuthUseCase.RevokeToken(c.Request.Context(), userIDStr, sessionID)
	if err != nil {
		response.HandleError(c, err, "Logout failed")
		return
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("access_token", "", -1, "/", "", false, true)
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
	userIDStr, ok := userID.(string)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid user context")
		return
	}

	err := h.AuthUseCase.RequestVerification(c.Request.Context(), userIDStr)
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
		h.log.WithContext(c.Request.Context()).Errorf("Register failed for user: %s\n with error: %v", req.Username, err)
		response.HandleError(c, err, "Register failed")
		return
	}

	// Set refresh token in HttpOnly cookie
	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "", false, true)
	c.SetCookie("access_token", res.AccessToken, int(res.ExpiresIn), "/", "", false, true)

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
	userID, exists := c.Get("user_id")
	if !exists || userID == nil {
		response.Unauthorized(c, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	username, ok := GetUsernameFromContext(c)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid username context")
		return
	}

	role, ok := GetRoleFromContext(c)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid role context")
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

// GetTicket godoc
// @Summary      Get WebSocket Ticket
// @Description  Generates a one-time ticket for WebSocket authentication.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/ticket [post]
func (h *AuthController) GetTicket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists || userID == nil {
		response.Unauthorized(c, exception.ErrUnauthorized, "user not authenticated")
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid user context")
		return
	}

	sessionID, ok := GetSessionIDFromContext(c)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid session context")
		return
	}

	role, ok := GetRoleFromContext(c)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid role context")
		return
	}

	username, ok := GetUsernameFromContext(c)
	if !ok {
		response.Unauthorized(c, exception.ErrUnauthorized, "invalid username context")
		return
	}

	orgID := c.Query("org_id")
	if orgID == "" {
		orgID = c.Query("organization_id")
	}

	// If orgID is still empty, try checking if user has a default context or require it depending on logic.
	// For now, we allow empty orgID if the ticket is just for connection, but usually we need context.
	// However, the TicketManager supports storing it.

	ticket, err := h.AuthUseCase.GetTicket(
		c.Request.Context(),
		model.UserSessionContext{
			UserID:    userIDStr,
			OrgID:     orgID,
			SessionID: sessionID,
			Role:      role,
			Username:  username,
		},
	)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate WebSocket ticket")
		response.InternalServerError(c, errors.New("failed to generate ticket"), "internal server error")
		return
	}

	response.Success(c, gin.H{
		"ticket":     ticket,
		"expires_in": 30, // seconds
	})
}

// Helper functions to get data from context safely (duplicated from middleware for now, or ensure middleware sets them)
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
