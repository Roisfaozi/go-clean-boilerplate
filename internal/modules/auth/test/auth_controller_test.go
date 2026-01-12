package test_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}
func newTestAuthController(mockUseCase *mocks.MockAuthUseCase) *authHandler.AuthController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	return authHandler.NewAuthController(mockUseCase, log, validator.New())
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	reqBody := model.LoginRequest{Username: "testuser", Password: "password123"}
	loginRes := &model.LoginResponse{AccessToken: "access_token", TokenType: "Bearer"}
	refreshToken := "refresh_token"

	mockUseCase.On("Login", mock.Anything, mock.MatchedBy(func(r model.LoginRequest) bool {
		return r.Username == reqBody.Username
	})).Return(loginRes, refreshToken, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token=refresh_token")
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	reqBody := model.ForgotPasswordRequest{Email: "test@example.com"}
	mockUseCase.On("ForgotPassword", mock.Anything, reqBody.Email).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["data"].(map[string]interface{})["message"], "reset link will be sent")
}

func TestAuthHandler_ForgotPassword_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	req, _ := http.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString(`{"email":`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "ForgotPassword", mock.Anything, mock.Anything)
}

func TestAuthHandler_ForgotPassword_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	reqBody := model.ForgotPasswordRequest{Email: "invalid-email"} // Invalid format
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "ForgotPassword", mock.Anything, mock.Anything)
}

func TestAuthHandler_ForgotPassword_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/forgot-password", handler.ForgotPassword)

	reqBody := model.ForgotPasswordRequest{Email: "test@example.com"}
	mockUseCase.On("ForgotPassword", mock.Anything, reqBody.Email).Return(errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/reset-password", handler.ResetPassword)

	reqBody := model.ResetPasswordRequest{
		Token:       "valid-token",
		NewPassword: "new-strong-password-123",
	}
	mockUseCase.On("ResetPassword", mock.Anything, reqBody.Token, reqBody.NewPassword).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ResetPassword_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/reset-password", handler.ResetPassword)

	req, _ := http.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBufferString(`{"token":`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_ResetPassword_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/reset-password", handler.ResetPassword)

	reqBody := model.ResetPasswordRequest{
		Token:       "",      // Empty token
		NewPassword: "short", // Too short
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_ResetPassword_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/reset-password", handler.ResetPassword)

	reqBody := model.ResetPasswordRequest{
		Token:       "invalid-token",
		NewPassword: "new-strong-password-123",
	}
	mockUseCase.On("ResetPassword", mock.Anything, reqBody.Token, reqBody.NewPassword).Return(usecase.ErrInvalidResetToken)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Controller uses HandleError which maps ErrInvalidResetToken (aliased to ErrBadRequest)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	userID := "user-123"
	sessionID := "session-abc"

	mockUseCase.On("RevokeToken", mock.Anything, userID, sessionID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/logout", nil)
	// The controller expects "user_id" and "session_id" (snake_case) not "userID" and "sessionID"
	c.Set("user_id", userID)
	c.Set("session_id", sessionID)

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token=")
}

func TestAuthHandler_Login_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	reqBody := model.LoginRequest{Username: "testuser", Password: "wrong-password"}
	mockUseCase.On("Login", mock.Anything, mock.MatchedBy(func(r model.LoginRequest) bool {
		return r.Username == reqBody.Username
	})).Return(nil, "", usecase.ErrInvalidCredentials)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// ErrInvalidCredentials maps to 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	refreshToken := "valid-refresh-token"
	newAccessToken := "new-access-token"
	newRefreshToken := "new-refresh-token"
	tokenResp := &model.TokenResponse{AccessToken: newAccessToken, TokenType: "Bearer"}

	mockUseCase.On("RefreshToken", mock.Anything, refreshToken).Return(tokenResp, newRefreshToken, nil)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token="+newRefreshToken)
}

func TestAuthHandler_RefreshToken_NoCookie(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
	// No cookie set

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertNotCalled(t, "RefreshToken", mock.Anything, mock.Anything)
}

func TestAuthHandler_RefreshToken_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	refreshToken := "invalid-token"
	mockUseCase.On("RefreshToken", mock.Anything, refreshToken).Return(nil, "", usecase.ErrInvalidToken)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// ErrInvalidToken maps to 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/logout", nil)
	// Missing user_id and session_id in context

	handler.Logout(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertNotCalled(t, "RevokeToken", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_Logout_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	userID := "user-123"
	sessionID := "session-abc"

	mockUseCase.On("RevokeToken", mock.Anything, userID, sessionID).Return(errors.New("redis error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/logout", nil)
	c.Set("user_id", userID)
	c.Set("session_id", sessionID)

	handler.Logout(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- EMAIL VERIFICATION HANDLER TESTS ---

func TestAuthHandler_VerifyEmail_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/verify-email", handler.VerifyEmail)

	reqBody := model.VerifyEmailRequest{Token: "valid-verification-token"}
	mockUseCase.On("VerifyEmail", mock.Anything, reqBody.Token).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["data"].(map[string]interface{})["message"], "verified successfully")
}

func TestAuthHandler_VerifyEmail_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/verify-email", handler.VerifyEmail)

	req, _ := http.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "VerifyEmail", mock.Anything, mock.Anything)
}

func TestAuthHandler_VerifyEmail_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/verify-email", handler.VerifyEmail)

	reqBody := model.VerifyEmailRequest{Token: ""} // Empty token
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "VerifyEmail", mock.Anything, mock.Anything)
}

func TestAuthHandler_VerifyEmail_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/verify-email", handler.VerifyEmail)

	reqBody := model.VerifyEmailRequest{Token: "invalid-token"}
	mockUseCase.On("VerifyEmail", mock.Anything, reqBody.Token).Return(usecase.ErrInvalidVerificationToken)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_ResendVerification_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	userID := "user-123"
	mockUseCase.On("RequestVerification", mock.Anything, userID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/resend-verification", nil)
	c.Set("user_id", userID)

	handler.ResendVerification(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_ResendVerification_Unauthenticated(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/resend-verification", nil)
	// Missing user_id in context

	handler.ResendVerification(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertNotCalled(t, "RequestVerification", mock.Anything, mock.Anything)
}

func TestAuthHandler_ResendVerification_AlreadyVerified(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthController(mockUseCase)

	userID := "user-123"
	mockUseCase.On("RequestVerification", mock.Anything, userID).Return(usecase.ErrAlreadyVerified)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/resend-verification", nil)
	c.Set("user_id", userID)

	handler.ResendVerification(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertExpectations(t)
}

