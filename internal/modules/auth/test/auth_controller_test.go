package test

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
	router := gin.Default()
	return router
}

func newTestAuthHandler(mockUseCase *mocks.MockAuthUseCase) *authHandler.AuthController {
	return authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	reqBody := model.LoginRequest{Username: "testuser", Password: "password123"}
	loginRes := &model.LoginResponse{AccessToken: "access_token", TokenType: "Bearer"}
	refreshToken := "refresh_token"

	mockUseCase.On("Login", mock.Anything, reqBody).Return(loginRes, refreshToken, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token=refresh_token")
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"username":`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "Login", mock.Anything, mock.Anything)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	reqBody := model.LoginRequest{Username: "", Password: "password123"}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "Login", mock.Anything, mock.Anything)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	reqBody := model.LoginRequest{Username: "wrong", Password: "wrong-password"}
	mockUseCase.On("Login", mock.Anything, reqBody).Return(nil, "", usecase.ErrInvalidCredentials)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_Login_UseCaseGenericError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)

	reqBody := model.LoginRequest{Username: "testuser", Password: "password123"}
	mockUseCase.On("Login", mock.Anything, reqBody).Return(nil, "", errors.New("something went wrong"))

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

// --- RefreshToken Handler Tests ---
func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	oldRefreshToken := "old_refresh_token"
	newAccessToken := &model.TokenResponse{AccessToken: "new_access_token", TokenType: "Bearer"}
	newRefreshToken := "new_refresh_token"

	mockUseCase.On("RefreshToken", mock.Anything, oldRefreshToken).Return(newAccessToken, newRefreshToken, nil)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: oldRefreshToken})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token=new_refresh_token")
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_NoCookie(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil) // No cookie

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertNotCalled(t, "RefreshToken", mock.Anything, mock.Anything)
}

func TestAuthHandler_RefreshToken_UseCaseInvalidTokenError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", handler.RefreshToken)

	oldRefreshToken := "invalid_refresh_token"
	mockUseCase.On("RefreshToken", mock.Anything, oldRefreshToken).Return(nil, "", usecase.ErrInvalidToken)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: oldRefreshToken})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertExpectations(t)
}

// --- Logout Handler Tests ---
func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/logout", handler.Logout)

	userID := "user-123"
	sessionID := "session-abc"

	mockUseCase.On("RevokeToken", mock.Anything, userID, sessionID).Return(nil)

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Set("session_id", sessionID)

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "refresh_token=")
	mockUseCase.AssertExpectations(t)
}

func TestAuthHandler_Logout_NoUserIDInContext(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/logout", handler.Logout)

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("session_id", "session-abc")

	handler.Logout(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUseCase.AssertNotCalled(t, "RevokeToken", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthHandler_Logout_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := newTestAuthHandler(mockUseCase)
	router := setupAuthTestRouter()
	router.POST("/auth/logout", handler.Logout)

	userID := "user-123"
	sessionID := "session-abc"
	mockUseCase.On("RevokeToken", mock.Anything, userID, sessionID).Return(errors.New("db error"))

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Set("session_id", sessionID)

	handler.Logout(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}
