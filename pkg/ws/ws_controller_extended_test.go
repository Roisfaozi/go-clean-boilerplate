package ws_test

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebSocketController_HandleWebSocket_WithUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup Mocks
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockUserRepo.On("FindByID", mock.Anything, "u1").Return(&entity.User{
		ID: "u1", Name: "User One", AvatarURL: "avatar.jpg",
	}, nil)

	// Use setupTestServer to get a manager (and server, which we won't use directly for routing)
	manager, server := setupTestServer(nil)
	defer server.Close()
	defer manager.Stop()

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	controller := ws.NewWebSocketController(logger, manager, []string{"*"}, mockUserRepo, nil)

	// Setup Router
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("user_id", "u1")
		c.Set("organization_id", "org1")
		controller.HandleWebSocket(c)
	})

	// Create Server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Connect
	wsURL := "ws" + ts.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Verify user data via Presence (indirectly)
	// We wait for client registration
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, manager.ClientCount())

	mockUserRepo.AssertExpectations(t)
}

func TestWebSocketController_HandleWebSocket_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup Mocks
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockUserRepo.On("FindByID", mock.Anything, "u1").Return(nil, errors.New("not found"))

	manager, server := setupTestServer(nil)
	defer server.Close()
	defer manager.Stop()

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	controller := ws.NewWebSocketController(logger, manager, []string{"*"}, mockUserRepo, nil)

	// Setup Router
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("user_id", "u1")
		controller.HandleWebSocket(c)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	wsURL := "ws" + ts.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, manager.ClientCount()) // Connection still succeeds

	mockUserRepo.AssertExpectations(t)
}
