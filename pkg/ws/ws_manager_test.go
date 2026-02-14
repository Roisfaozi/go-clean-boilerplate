package ws_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

type NoOpPresenceManager struct{}

func (m *NoOpPresenceManager) SetUserOnline(ctx context.Context, orgID, userID string, userData *ws.PresenceUser) error {
	return nil
}
func (m *NoOpPresenceManager) SetUserOffline(ctx context.Context, orgID, userID string) error {
	return nil
}
func (m *NoOpPresenceManager) GetOnlineUsers(ctx context.Context, orgID string) ([]ws.PresenceUser, error) {
	return []ws.PresenceUser{}, nil
}
func (m *NoOpPresenceManager) RefreshUserHeartbeat(ctx context.Context, orgID, userID string) error {
	return nil
}
func (m *NoOpPresenceManager) PruneStaleUsers(ctx context.Context, timeout time.Duration) (map[string][]string, error) {
	return nil, nil
}

func setupTestServer() (*ws.WebSocketManager, *httptest.Server) {
	config := &ws.WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     54 * time.Second,
		MaxMessageSize: 512 * 1024,
	}
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})
	// For unit tests, we don't need Redis scaling, so pass nil
	presence := &NoOpPresenceManager{}
	manager := ws.NewWebSocketManager(config, logger, nil, presence)
	go manager.Run()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := ws.NewWebsocketClient(conn, manager, logger, config, "u1", "org1", nil)
		manager.RegisterClient(client)
		go client.WritePump()
		go client.ReadPump()
	})

	server := httptest.NewServer(handler)
	return manager, server
}

func connectClient(url string) (*websocket.Conn, error) {
	wsURL := "ws" + strings.TrimPrefix(url, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	return conn, err
}

func waitForMessage(conn *websocket.Conn, msgType string, channel string) (*ws.ServerMessage, error) {
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil { // Check error
		return nil, err
	}
	for {
		var msg ws.ServerMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			return nil, err
		}
		if msg.Type == msgType && (channel == "" || msg.Channel == channel) {
			return &msg, nil
		}
	}
}

func TestNewWebSocketManager(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()
	assert.NotNil(t, manager)
}

func TestWebSocketManager_Integration(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	conn, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }() // Ignore close error

	// Wait for registration
	for i := 0; i < 10; i++ {
		if manager.ClientCount() == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 1, manager.ClientCount())

	// Subscribe
	err = conn.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "test-channel"})
	require.NoError(t, err)

	// Wait for subscription info
	_, err = waitForMessage(conn, "info", "test-channel")
	require.NoError(t, err)

	// Wait for subscription to be processed by manager
	for i := 0; i < 20; i++ {
		if manager.GetChannelClients("test-channel") == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 1, manager.GetChannelClients("test-channel"))

	// Broadcast - Must send a message structure that client expects (ws.ServerMessage)
	// to pass waitForMessage check
	broadcastContent := ws.ServerMessage{
		Type:    "message",
		Channel: "test-channel",
		Data:    map[string]string{"event": "hello"},
	}
	broadcastBytes, _ := json.Marshal(broadcastContent) // No error check needed for marshal in test
	manager.BroadcastToChannel("test-channel", broadcastBytes)

	// Wait for broadcast message
	msg, err := waitForMessage(conn, "message", "test-channel")
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "message", msg.Type)
	assert.Equal(t, "test-channel", msg.Channel)
}

func TestBroadcastToChannel(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	// Client 1 -> channel1
	c1, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }() // Ignore error
	require.NoError(t, c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "channel1"}))
	_, err = waitForMessage(c1, "info", "channel1")
	require.NoError(t, err)

	// Client 2 -> channel1
	c2, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c2.Close() }() // Ignore error
	require.NoError(t, c2.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "channel1"}))
	_, err = waitForMessage(c2, "info", "channel1")
	require.NoError(t, err)

	// Client 3 -> channel2
	c3, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c3.Close() }() // Ignore error
	require.NoError(t, c3.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "channel2"}))
	_, err = waitForMessage(c3, "info", "channel2")
	require.NoError(t, err)

	// Wait for subscriptions to be processed
	for i := 0; i < 20; i++ {
		if manager.GetChannelClients("channel1") == 2 && manager.GetChannelClients("channel2") == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 2, manager.GetChannelClients("channel1"), "Should have 2 clients in channel1")
	assert.Equal(t, 1, manager.GetChannelClients("channel2"), "Should have 1 client in channel2")

	// Broadcast to channel1
	broadcastContent := ws.ServerMessage{
		Type:    "message",
		Channel: "channel1",
		Data:    map[string]string{"msg": "hello channel 1"},
	}
	msgBytes, _ := json.Marshal(broadcastContent)
	manager.BroadcastToChannel("channel1", msgBytes)

	// Verify c1 received
	_, err = waitForMessage(c1, "message", "channel1")
	assert.NoError(t, err)

	// Verify c2 received
	_, err = waitForMessage(c2, "message", "channel1")
	assert.NoError(t, err)

	// Verify c3 did NOT receive
	if err := c3.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		t.Fatalf("Failed to set read deadline for c3: %v", err)
	}
	var msg ws.ServerMessage
	err = c3.ReadJSON(&msg)
	assert.Error(t, err) // Should timeout or EOF
}
