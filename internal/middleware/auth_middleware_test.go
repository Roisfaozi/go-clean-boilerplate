package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

type MockTicketManager struct {
	mock.Mock
}

func (m *MockTicketManager) CreateTicket(ctx context.Context, userID, orgID, sessionID, role, username string) (string, error) {
	args := m.Called(ctx, userID, orgID, sessionID, role, username)
	return args.String(0), args.Error(1)
}

func (m *MockTicketManager) ValidateTicket(ctx context.Context, ticket string) (*ws.UserContext, error) {
	args := m.Called(ctx, ticket)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ws.UserContext), args.Error(1)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	claims := &jwt.Claims{
		UserID:    "user123",
		SessionID: "session456",
		Role:      "role:user",
		Username:  "testuser",
	}

	mockAuthUseCase.On("ValidateAccessToken", "valid_token").Return(claims, nil)
	mockAuthUseCase.On("Verify", mock.Anything, claims.UserID, claims.SessionID).Return(&model.Auth{ID: claims.SessionID}, nil)

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, claims.UserID, c.GetString("user_id"))
	assert.Equal(t, claims.SessionID, c.GetString("session_id"))
	assert.Equal(t, claims.Role, c.GetString("user_role"))
	assert.Equal(t, claims.Username, c.GetString("username"))
	mockAuthUseCase.AssertExpectations(t)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockAuthUseCase.AssertNotCalled(t, "ValidateAccessToken")
	mockAuthUseCase.AssertNotCalled(t, "Verify")
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidToken")
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockAuthUseCase.AssertNotCalled(t, "ValidateAccessToken")
	mockAuthUseCase.AssertNotCalled(t, "Verify")
}

func TestAuthMiddleware_InvalidTokenSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.signature.token")
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockAuthUseCase.On("ValidateAccessToken", "invalid.signature.token").Return(nil, errors.New("invalid signature"))

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockAuthUseCase.AssertExpectations(t)
	mockAuthUseCase.AssertNotCalled(t, "Verify")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer expired_token")
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockAuthUseCase.On("ValidateAccessToken", "expired_token").Return(nil, errors.New("token is expired"))

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockAuthUseCase.AssertExpectations(t)
	mockAuthUseCase.AssertNotCalled(t, "Verify")
}

func TestAuthMiddleware_SessionRevoked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	claims := &jwt.Claims{
		UserID:    "user123",
		SessionID: "session456",
		Role:      "role:user",
		Username:  "testuser",
	}

	mockAuthUseCase.On("ValidateAccessToken", "valid_token").Return(claims, nil)
	mockAuthUseCase.On("Verify", mock.Anything, claims.UserID, claims.SessionID).Return(nil, nil) // Return nil session = revoked

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockAuthUseCase.AssertExpectations(t)
}

func TestAuthMiddleware_SessionVerifyError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	c.Request = req

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	claims := &jwt.Claims{
		UserID:    "user123",
		SessionID: "session456",
		Role:      "role:user",
		Username:  "testuser",
	}

	mockAuthUseCase.On("ValidateAccessToken", "valid_token").Return(claims, nil)
	mockAuthUseCase.On("Verify", mock.Anything, claims.UserID, claims.SessionID).Return(nil, errors.New("database error"))

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	authMiddleware.ValidateToken()(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
	mockAuthUseCase.AssertExpectations(t)
}

func TestAuthMiddleware_ContextSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthUseCase := new(authMocks.MockAuthUseCase)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	claims := &jwt.Claims{
		UserID:    "user123",
		SessionID: "session456",
		Role:      "role:admin",
		Username:  "adminuser",
	}

	mockAuthUseCase.On("ValidateAccessToken", "valid_token").Return(claims, nil)
	mockAuthUseCase.On("Verify", mock.Anything, claims.UserID, claims.SessionID).Return(&model.Auth{ID: claims.SessionID}, nil)

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(mockAuthUseCase, logger, mockTicketManager)

	r := gin.New()
	r.Use(authMiddleware.ValidateToken())

	r.GET("/test", func(c *gin.Context) {
		assert.Equal(t, claims.UserID, c.GetString("user_id"))
		assert.Equal(t, claims.SessionID, c.GetString("session_id"))
		assert.Equal(t, claims.Role, c.GetString("user_role"))
		assert.Equal(t, claims.Username, c.GetString("username"))
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthUseCase.AssertExpectations(t)
}
