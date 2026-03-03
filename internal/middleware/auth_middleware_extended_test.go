package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthMiddleware_ValidateWebSocketToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/ws?ticket=valid_ticket", nil)
	c.Request = req

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	userCtx := &ws.UserContext{
		UserID:         "user123",
		SessionID:      "session456",
		Role:           "role:user",
		Username:       "testuser",
		OrganizationID: "org789",
	}

	mockTicketManager := new(MockTicketManager)
	mockTicketManager.On("ValidateTicket", mock.Anything, "valid_ticket").Return(userCtx, nil)

	authMiddleware := middleware.NewAuthMiddleware(nil, logger, mockTicketManager)
	authMiddleware.ValidateWebSocketToken()(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, userCtx.UserID, c.GetString("user_id"))
	assert.Equal(t, userCtx.SessionID, c.GetString("session_id"))
	assert.Equal(t, userCtx.Role, c.GetString("user_role"))
	assert.Equal(t, userCtx.Username, c.GetString("username"))
	assert.Equal(t, userCtx.OrganizationID, c.GetString("organization_id"))
	mockTicketManager.AssertExpectations(t)
}

func TestAuthMiddleware_ValidateWebSocketToken_NoTicket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/ws", nil)
	c.Request = req

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockTicketManager := new(MockTicketManager)
	authMiddleware := middleware.NewAuthMiddleware(nil, logger, mockTicketManager)
	authMiddleware.ValidateWebSocketToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockTicketManager.AssertNotCalled(t, "ValidateTicket")
}

func TestAuthMiddleware_ValidateWebSocketToken_InvalidTicket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/ws?ticket=invalid_ticket", nil)
	c.Request = req

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockTicketManager := new(MockTicketManager)
	mockTicketManager.On("ValidateTicket", mock.Anything, "invalid_ticket").Return(nil, errors.New("invalid ticket"))

	authMiddleware := middleware.NewAuthMiddleware(nil, logger, mockTicketManager)
	authMiddleware.ValidateWebSocketToken()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockTicketManager.AssertExpectations(t)
}

func TestGetUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exists and valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", "user123")

		val, ok := middleware.GetUserIDFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "user123", val)
	})

	t.Run("not exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		val, ok := middleware.GetUserIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", 123)

		val, ok := middleware.GetUserIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("empty string", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", "")

		val, ok := middleware.GetUserIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestAuthMiddleware_GetSessionIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exists and valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("session_id", "session123")

		val, ok := middleware.GetSessionIDFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "session123", val)
	})

	t.Run("not exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		val, ok := middleware.GetSessionIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("session_id", 123)

		val, ok := middleware.GetSessionIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("empty string", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("session_id", "")

		val, ok := middleware.GetSessionIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestAuthMiddleware_GetRoleFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exists and valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_role", "admin")

		val, ok := middleware.GetRoleFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "admin", val)
	})

	t.Run("not exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		val, ok := middleware.GetRoleFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_role", 123)

		val, ok := middleware.GetRoleFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("empty string", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_role", "")

		val, ok := middleware.GetRoleFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestAuthMiddleware_GetUsernameFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exists and valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("username", "testuser")

		val, ok := middleware.GetUsernameFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "testuser", val)
	})

	t.Run("not exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		val, ok := middleware.GetUsernameFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("username", 123)

		val, ok := middleware.GetUsernameFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("empty string", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("username", "")

		val, ok := middleware.GetUsernameFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}
