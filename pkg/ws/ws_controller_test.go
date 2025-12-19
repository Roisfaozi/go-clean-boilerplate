package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockManager is a mock implementation of the Manager interface
type MockManager struct {
	mock.Mock
}

func (m *MockManager) Run() {
	m.Called()
}

func (m *MockManager) RegisterClient(client *Client) {
	m.Called(client)
}

func (m *MockManager) UnregisterClient(client *Client) {
	m.Called(client)
}

func (m *MockManager) BroadcastToChannel(channel string, message []byte) {
	m.Called(channel, message)
}

func (m *MockManager) SubscribeToChannel(client *Client, channel string) {
	m.Called(client, channel)
}

func (m *MockManager) UnsubscribeFromChannel(client *Client, channel string) {
	m.Called(client, channel)
}

func (m *MockManager) GetChannelClients(channel string) int {
	args := m.Called(channel)
	return args.Int(0)
}

func TestCheckOrigin(t *testing.T) {
	// Setup generic logger
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // Silence logs during test

	// Mock manager
	mockManager := new(MockManager)

	// Set generic allowed origins
	allowedOrigins := []string{"http://localhost:3000", "https://mydomain.com"}

	controller := NewWebSocketController(logger, mockManager, allowedOrigins)

	// Create a test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", controller.HandleWebSocket)

	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	t.Run("Allowed Origin Should Connect", func(t *testing.T) {
		header := http.Header{}
		header.Add("Origin", "http://localhost:3000")

		// Expect RegisterClient to be called if connection succeeds
		mockManager.On("RegisterClient", mock.Anything).Return().Once()
		mockManager.On("UnregisterClient", mock.Anything).Return().Maybe()

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
		assert.NoError(t, err)
		if conn != nil {
			conn.Close()
		}
	})

	t.Run("Disallowed Origin Should Fail", func(t *testing.T) {
		header := http.Header{}
		header.Add("Origin", "http://evil.com")

		_, _, err := websocket.DefaultDialer.Dial(wsURL, header)

		assert.Error(t, err)
		// Usually returns 403 Forbidden
		assert.Contains(t, err.Error(), "bad handshake")
	})

	t.Run("Wildcard Origin Should Allow All", func(t *testing.T) {
		wildcardController := NewWebSocketController(logger, mockManager, []string{"*"})
		wildcardRouter := gin.New()
		wildcardRouter.GET("/ws", wildcardController.HandleWebSocket)

		wildcardServer := httptest.NewServer(wildcardRouter)
		defer wildcardServer.Close()

		wildcardWsURL := "ws" + strings.TrimPrefix(wildcardServer.URL, "http") + "/ws"

		header := http.Header{}
		header.Add("Origin", "http://anywhere.com")

		mockManager.On("RegisterClient", mock.Anything).Return().Once()
		mockManager.On("UnregisterClient", mock.Anything).Return().Maybe()

		conn, _, err := websocket.DefaultDialer.Dial(wildcardWsURL, header)
		assert.NoError(t, err)
		if conn != nil {
			conn.Close()
		}
	})
}
