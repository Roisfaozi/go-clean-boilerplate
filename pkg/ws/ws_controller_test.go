package ws

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// NoOpLogWriter suppresses log output for tests
type NoOpLogWriter struct{}

func (w *NoOpLogWriter) Write([]byte) (int, error) {
	return 0, nil
}

// MockManager is a mock implementation of the Manager interface
type MockManager struct {
	mock.Mock
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

func (m *MockManager) Run() {
	m.Called()
}

func (m *MockManager) PresenceUpdate(orgID string, event string, userData *PresenceUser) {
	m.Called(orgID, event, userData)
}

func (m *MockManager) GetPresenceManager() PresenceManager {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(PresenceManager)
}

// TestNewWebSocketController_CheckOrigin is a placeholder for the logic test, but verification is done in TestWebSocketOrigin

// Simplified test checking the logic by making `upgrader` public or adding a getter?
// No, I shouldn't change the code just for testing if I can avoid it.
// I can verify the logic by running a real test server.

func TestWebSocketOrigin(t *testing.T) {
	log := logrus.New()
	log.SetOutput(&NoOpLogWriter{})

	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		shouldConnect  bool
	}{
		{
			name:           "Allowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://example.com",
			shouldConnect:  true,
		},
		{
			name:           "Disallowed Origin",
			allowedOrigins: []string{"http://example.com"},
			requestOrigin:  "http://malicious.com",
			shouldConnect:  false,
		},
		{
			name:           "Wildcard Origin",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://anywhere.com",
			shouldConnect:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mock for each subtest
			manager := new(MockManager)
			manager.On("RegisterClient", mock.Anything).Return().Maybe()
			manager.On("UnregisterClient", mock.Anything).Return().Maybe()

			ctrl := NewWebSocketController(log, manager, tt.allowedOrigins, nil, nil)

			// Start a test server
			r := http.NewServeMux()
			r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
				// We need to convert http.Request to Gin context?
				// Or we can just manually invoke the upgrader if we could access it.
				// But `ctrl.HandleWebSocket` expects `*gin.Context`.

				// Let's create a Gin engine and pass the request to it.
				gin.SetMode(gin.ReleaseMode)
				engine := gin.New()
				engine.GET("/ws", ctrl.HandleWebSocket)
				engine.ServeHTTP(w, r)
			})

			server := httptest.NewServer(r)
			defer server.Close()

			// Build WebSocket URL
			wsURL := "ws" + server.URL[4:] + "/ws"

			header := http.Header{}
			if tt.requestOrigin != "" {
				header.Set("Origin", tt.requestOrigin)
			}

			conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)

			if tt.shouldConnect {
				assert.NoError(t, err, "Should connect")
				if err == nil {
					assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
					// Close connection to prevent dangling goroutines
					_ = conn.Close()
				}
			} else {
				assert.Error(t, err, "Should fail to connect")
				if resp != nil {
					assert.Equal(t, http.StatusForbidden, resp.StatusCode)
				}
			}
		})
	}
}
