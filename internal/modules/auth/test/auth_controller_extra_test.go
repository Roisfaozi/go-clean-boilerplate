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

func TestAuthHandler_Login_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockAuthUseCase)
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())
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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())

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
	handler := authHandler.NewAuthController(mockUseCase, logrus.New(), validator.New())

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
