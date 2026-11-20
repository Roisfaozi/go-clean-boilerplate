package http

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const ssoStateCookieName = "sso_state"

type CookieConfig struct {
	Domain   string
	SameSite string
	Secure   *bool
}

type AuthController struct {
	AuthUseCase usecase.AuthUseCase
	log         *logrus.Logger
	validate    *validator.Validate
	cookieCfg   CookieConfig
}

func NewAuthController(useCase usecase.AuthUseCase, log *logrus.Logger, validate *validator.Validate, cookieCfg CookieConfig) *AuthController {
	return &AuthController{
		AuthUseCase: useCase,
		log:         log,
		validate:    validate,
		cookieCfg:   cookieCfg,
	}
}

func (h *AuthController) sameSite() http.SameSite {
	switch strings.ToLower(h.cookieCfg.SameSite) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func generateSSOState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (h *AuthController) resolveSecureHTTP(r *http.Request) bool {
	if h.cookieCfg.Secure != nil {
		return *h.cookieCfg.Secure
	}
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func (h *AuthController) setAuthCookieHTTP(w http.ResponseWriter, r *http.Request, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.resolveSecureHTTP(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthController) setCsrfCookieHTTP(w http.ResponseWriter, r *http.Request, maxAge int) {
	csrfToken := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		MaxAge:   maxAge,
		HttpOnly: false,
		Secure:   h.resolveSecureHTTP(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthController) clearCsrfCookieHTTP(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		Domain:   h.cookieCfg.Domain,
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   h.resolveSecureHTTP(r),
		SameSite: h.sameSite(),
	})
}

func (h *AuthController) clearAuthCookieHTTP(w http.ResponseWriter, r *http.Request, name string) {
	h.setAuthCookieHTTP(w, r, name, "", -1)
}

func setStateCookieHTTP(w http.ResponseWriter, r *http.Request, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

func clearStateCookieHTTP(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

// HTTPLogin godoc
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
func (h *AuthController) HTTPLogin(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.log.WithContext(r.Context()).WithError(err).Error("Login failed: could not bind request")
		response.WriteHTTPError(w, exception.ErrBadRequest, "could not bind request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.log.WithContext(r.Context()).WithError(err).Error("Login failed: validation error")
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}

	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()

	res, refreshToken, err := h.AuthUseCase.Login(r.Context(), req)
	if err != nil {
		h.log.WithContext(r.Context()).Errorf("Login failed for user: %s", req.Username)
		response.WriteHTTPError(w, err, "Login failed")
		return
	}

	h.setAuthCookieHTTP(w, r, "refresh_token", refreshToken, int(h.AuthUseCase.GetRefreshTokenDuration().Seconds()))
	h.setAuthCookieHTTP(w, r, "access_token", res.AccessToken, int(res.ExpiresIn))
	h.setCsrfCookieHTTP(w, r, int(h.AuthUseCase.GetRefreshTokenDuration().Seconds()))

	response.WriteSuccess(w, http.StatusOK, res)
}

// HTTPRefreshToken godoc
// @Summary      Refresh access token
// @Description  Refreshes access and refresh tokens using the refresh token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerTokenResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/refresh [post]
func (h *AuthController) HTTPRefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		h.log.WithContext(r.Context()).Warn("Refresh token not found in cookie")
		response.WriteHTTPError(w, exception.ErrUnauthorized, "refresh token not found")
		return
	}

	res, newRefreshToken, err := h.AuthUseCase.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		response.WriteHTTPError(w, err, "Refresh token failed")
		return
	}

	h.setAuthCookieHTTP(w, r, "refresh_token", newRefreshToken, int(h.AuthUseCase.GetRefreshTokenDuration().Seconds()))
	h.setAuthCookieHTTP(w, r, "access_token", res.AccessToken, int(res.ExpiresIn))
	h.setCsrfCookieHTTP(w, r, int(h.AuthUseCase.GetRefreshTokenDuration().Seconds()))
	response.WriteSuccess(w, http.StatusOK, res)
}

// HTTPLogout godoc
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
func (h *AuthController) HTTPLogout(w http.ResponseWriter, r *http.Request) {
	userID, okUserID := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	sessionID, okSessionID := delivery.GetContextString(r.Context(), delivery.SessionIDKey)

	if !okUserID || !okSessionID || userID == "" || sessionID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	err := h.AuthUseCase.RevokeToken(r.Context(), userID, sessionID)
	if err != nil {
		response.WriteHTTPError(w, err, "Logout failed")
		return
	}

	h.clearAuthCookieHTTP(w, r, "refresh_token")
	h.clearAuthCookieHTTP(w, r, "access_token")
	h.clearCsrfCookieHTTP(w, r)
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// HTTPForgotPassword godoc
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
func (h *AuthController) HTTPForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ForgotPasswordRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}

	err := h.AuthUseCase.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to process forgot password request")
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "If the email is registered, a reset link will be sent shortly."})
}

// HTTPResetPassword godoc
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
func (h *AuthController) HTTPResetPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ResetPasswordRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}

	err := h.AuthUseCase.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to reset password")
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

// HTTPVerifyEmail godoc
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
func (h *AuthController) HTTPVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req model.VerifyEmailRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		response.WriteHTTPError(w, exception.ErrBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}

	err := h.AuthUseCase.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to verify email")
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

// HTTPResendVerification godoc
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
func (h *AuthController) HTTPResendVerification(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	err := h.AuthUseCase.RequestVerification(r.Context(), userID)
	if err != nil {
		response.WriteHTTPError(w, err, "failed to request verification email")
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "verification email sent successfully"})
}

// HTTPRegister godoc
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
func (h *AuthController) HTTPRegister(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := response.DecodeJSON(r, &req, 1024*1024); err != nil {
		h.log.WithContext(r.Context()).WithError(err).Error("Register failed: could not bind request")
		response.WriteHTTPError(w, exception.ErrBadRequest, "could not bind request")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.log.WithContext(r.Context()).WithError(err).Error("Register failed: validation error")
		response.WriteHTTPError(w, exception.ErrValidationError, validation.FormatValidationErrors(err))
		return
	}

	req.IPAddress = r.RemoteAddr
	req.UserAgent = r.UserAgent()

	res, refreshToken, err := h.AuthUseCase.Register(r.Context(), req)
	if err != nil {
		h.log.WithContext(r.Context()).Errorf("Register failed for user: %s\n with error: %v", req.Username, err)
		response.WriteHTTPError(w, err, "Register failed")
		return
	}

	h.setAuthCookieHTTP(w, r, "refresh_token", refreshToken, int(h.AuthUseCase.GetRefreshTokenDuration().Seconds()))
	h.setAuthCookieHTTP(w, r, "access_token", res.AccessToken, int(res.ExpiresIn))
	h.setCsrfCookieHTTP(w, r, int(h.AuthUseCase.GetRefreshTokenDuration().Seconds()))

	response.WriteSuccess(w, http.StatusCreated, res)
}

// HTTPMe godoc
// @Summary      Get current user
// @Description  Returns the currently authenticated user's information.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Router       /auth/me [get]
func (h *AuthController) HTTPMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	username, _ := delivery.GetContextString(r.Context(), delivery.UsernameKey)
	role, _ := delivery.GetContextString(r.Context(), delivery.RoleKey)

	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"user": map[string]any{
			"id":       userID,
			"username": username,
			"role":     role,
		},
	})
}

// HTTPGetTicket godoc
// @Summary      Get WebSocket Ticket
// @Description  Generates a one-time ticket for WebSocket authentication.
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        org_id query string false "Organization ID"
// @Param        organization_id query string false "Alternative Organization ID"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper "Unauthorized"
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper "Internal server error"
// @Router       /auth/ticket [post]
func (h *AuthController) HTTPGetTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
	if !ok || userID == "" {
		response.WriteHTTPError(w, exception.ErrUnauthorized, "user not authenticated")
		return
	}

	sessionID, _ := delivery.GetContextString(r.Context(), delivery.SessionIDKey)
	role, _ := delivery.GetContextString(r.Context(), delivery.RoleKey)
	username, _ := delivery.GetContextString(r.Context(), delivery.UsernameKey)

	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		orgID = r.URL.Query().Get("organization_id")
	}

	ticket, err := h.AuthUseCase.GetTicket(
		r.Context(),
		model.UserSessionContext{
			UserID:    userID,
			OrgID:     orgID,
			SessionID: sessionID,
			Role:      role,
			Username:  username,
		},
	)
	if err != nil {
		h.log.WithError(err).Error("Failed to generate WebSocket ticket")
		response.WriteHTTPError(w, errors.New("failed to generate ticket"), "internal server error")
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"ticket":     ticket,
		"expires_in": 30,
	})
}

// HTTPSSOLogin godoc
// @Summary      Initiate SSO Login
// @Description  Redirects the user to the specific OAuth2 provider (google, microsoft).
// @Tags         auth
// @Param        provider path string true "Provider Name" Enums(google, microsoft, github)
// @Success      302
// @Router       /auth/sso/{provider} [get]
func (ac *AuthController) HTTPSSOLogin(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	state, err := generateSSOState()
	if err != nil {
		response.WriteHTTPError(w, err, "failed to generate SSO state")
		return
	}

	url, err := ac.AuthUseCase.GetSSORedirectURL(r.Context(), provider, state)
	if err != nil {
		response.WriteHTTPError(w, err, "Failed to initiate SSO")
		return
	}

	setStateCookieHTTP(w, r, ssoStateCookieName, state, 300)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HTTPSSOCallback godoc
// @Summary      SSO Callback Handler
// @Description  Handles the OAuth2 callback, exchanges code for token, and authenticates user.
// @Tags         auth
// @Param        provider path string true "Provider Name" Enums(google, microsoft, github)
// @Param        code query string true "Authorization Code"
// @Param        state query string false "State"
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper{data=model.LoginResponse}
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Router       /auth/sso/{provider}/callback [get]
func (ac *AuthController) HTTPSSOCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		response.WriteHTTPError(w, exception.ErrBadRequest, "authorization code is required")
		return
	}

	cookie, err := r.Cookie(ssoStateCookieName)
	if err != nil || cookie.Value == "" || state == "" || state != cookie.Value {
		clearStateCookieHTTP(w, r, ssoStateCookieName)
		response.WriteHTTPError(w, exception.ErrUnauthorized, "invalid SSO state")
		return
	}

	clearStateCookieHTTP(w, r, ssoStateCookieName)
	res, refreshToken, err := ac.AuthUseCase.HandleSSOCallback(r.Context(), provider, code)
	if err != nil {
		response.WriteHTTPError(w, err, "Failed to handle SSO callback")
		return
	}

	ac.setAuthCookieHTTP(w, r, "refresh_token", refreshToken, int(ac.AuthUseCase.GetRefreshTokenDuration().Seconds()))
	ac.setAuthCookieHTTP(w, r, "access_token", res.AccessToken, int(res.ExpiresIn))
	ac.setCsrfCookieHTTP(w, r, int(ac.AuthUseCase.GetRefreshTokenDuration().Seconds()))

	response.WriteSuccess(w, http.StatusOK, res)
}
