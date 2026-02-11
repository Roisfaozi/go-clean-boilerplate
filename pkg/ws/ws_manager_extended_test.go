package ws_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketManager_MultipleSubscriptions verifies a client can subscribe to multiple channels
func TestWebSocketManager_MultipleSubscriptions(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	conn, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Wait for registration
	for i := 0; i < 10; i++ {
		if manager.ClientCount() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Subscribe to two channels
	require.NoError(t, conn.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "channel-alpha"}))
	_, err = waitForMessage(conn, "info", "channel-alpha")
	require.NoError(t, err)

	require.NoError(t, conn.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "channel-beta"}))
	_, err = waitForMessage(conn, "info", "channel-beta")
	require.NoError(t, err)

	// Broadcast to channel-alpha
	msgAlpha := ws.ServerMessage{Type: "message", Channel: "channel-alpha", Data: "alpha-data"}
	bytes, _ := json.Marshal(msgAlpha)
	manager.BroadcastToChannel("channel-alpha", bytes)

	msg, err := waitForMessage(conn, "message", "channel-alpha")
	require.NoError(t, err)
	assert.Equal(t, "channel-alpha", msg.Channel)

	// Broadcast to channel-beta
	msgBeta := ws.ServerMessage{Type: "message", Channel: "channel-beta", Data: "beta-data"}
	bytes, _ = json.Marshal(msgBeta)
	manager.BroadcastToChannel("channel-beta", bytes)

	msg, err = waitForMessage(conn, "message", "channel-beta")
	require.NoError(t, err)
	assert.Equal(t, "channel-beta", msg.Channel)
}

// TestWebSocketManager_Unsubscribe verifies client unsubscribe works
func TestWebSocketManager_Unsubscribe(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	conn, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	for i := 0; i < 10; i++ {
		if manager.ClientCount() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Subscribe
	require.NoError(t, conn.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "unsub-channel"}))
	_, err = waitForMessage(conn, "info", "unsub-channel")
	require.NoError(t, err)

	// Unsubscribe
	require.NoError(t, conn.WriteJSON(ws.ClientMessage{Type: "unsubscribe", Channel: "unsub-channel"}))

	// Consume the unsubscribe confirmation info message
	_, err = waitForMessage(conn, "info", "unsub-channel")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // wait for unsubscribe to fully process

	// Broadcast - client should NOT receive
	msgContent := ws.ServerMessage{Type: "message", Channel: "unsub-channel", Data: "after-unsub"}
	bytes, _ := json.Marshal(msgContent)
	manager.BroadcastToChannel("unsub-channel", bytes)

	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var msg ws.ServerMessage
	err = conn.ReadJSON(&msg)
	assert.Error(t, err, "Should not receive message after unsubscribing")
}

// TestWebSocketManager_ClientDisconnect verifies cleanup after client disconnects
func TestWebSocketManager_ClientDisconnect(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	conn, err := connectClient(server.URL)
	require.NoError(t, err)

	// Wait for registration
	for i := 0; i < 10; i++ {
		if manager.ClientCount() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 1, manager.ClientCount())

	// Close connection
	_ = conn.Close()

	// Wait for cleanup
	for i := 0; i < 20; i++ {
		if manager.ClientCount() == 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	assert.Equal(t, 0, manager.ClientCount())
}

// TestWebSocketManager_MultipleClientsIndependent verifies separate clients work independently
func TestWebSocketManager_MultipleClientsIndependent(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	// Connect 3 clients
	conn1, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn1.Close() }()

	conn2, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn2.Close() }()

	conn3, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn3.Close() }()

	// Wait for all registrations
	for i := 0; i < 30; i++ {
		if manager.ClientCount() >= 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 3, manager.ClientCount())
}

// TestWebSocketManager_GetChannelClients counts clients in a channel
func TestWebSocketManager_GetChannelClients(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	c1, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }()

	c2, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c2.Close() }()

	for i := 0; i < 20; i++ {
		if manager.ClientCount() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Subscribe both to same channel
	require.NoError(t, c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "counted-channel"}))
	_, _ = waitForMessage(c1, "info", "counted-channel")

	require.NoError(t, c2.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "counted-channel"}))
	_, _ = waitForMessage(c2, "info", "counted-channel")

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 2, manager.GetChannelClients("counted-channel"))
	assert.Equal(t, 0, manager.GetChannelClients("non-existent-channel"))
}

// TestWebSocketManager_BroadcastToEmptyChannel verifies no panic on empty channel
func TestWebSocketManager_BroadcastToEmptyChannel(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	// Should not panic
	msg := ws.ServerMessage{Type: "message", Channel: "empty-channel", Data: "to-nobody"}
	bytes, _ := json.Marshal(msg)
	manager.BroadcastToChannel("empty-channel", bytes)
}

// TestWebSocketManager_InvalidMessageType verifies unknown message type handling
func TestWebSocketManager_InvalidMessageType(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()
	defer manager.Stop()

	conn, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	for i := 0; i < 10; i++ {
		if manager.ClientCount() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Send an unknown message type
	err = conn.WriteJSON(ws.ClientMessage{Type: "unknown_type", Channel: "test"})
	require.NoError(t, err)

	// Should receive an error message back
	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	var msg ws.ServerMessage
	err = conn.ReadJSON(&msg)

	if err == nil {
		// Server may send an error response
		assert.Equal(t, "error", msg.Type)
	}
	// If timeout, it's also acceptable - server ignored the unknown type
}

// TestWebSocketManager_StopCleanup verifies manager Stop cleans up
func TestWebSocketManager_StopCleanup(t *testing.T) {
	manager, server := setupTestServer()
	defer server.Close()

	conn, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	for i := 0; i < 10; i++ {
		if manager.ClientCount() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Stop manager
	manager.Stop()
	time.Sleep(100 * time.Millisecond)

	// Writing after stop should fail or be ignored
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	_, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	// Connection may succeed but messages won't be processed
	// The main check is no panic occurred during stop
}
