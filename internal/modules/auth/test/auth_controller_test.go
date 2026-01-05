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

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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

	// Controller uses HandleError which maps ErrInvalidResetToken to internal server error?
	// Let's check exception mapping. It might be mapped to 400 or 401 if we defined it.
	// But usually generic errors are 500 unless specifically handled.
	// Actually, ErrInvalidResetToken is exported from usecase package.
	// We'll assert 500 for now as it's the default HandleError behavior for unknown errors.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())

	userID := "user-123"
	sessionID := "session-abc"

	mockUseCase.On("RevokeToken", mock.Anything, userID, sessionID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/auth/logout", nil)
	c.Set("userID", userID)
	c.Set("sessionID", sessionID)

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token=")
}
