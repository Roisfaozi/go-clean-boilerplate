//go:build e2e
// +build e2e

package realtime

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketE2E_NotificationFlow(t *testing.T) {
	// 1. Setup Server
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	// 2. Establish WebSocket Connection
	// Convert http:// to ws://
	wsURL := strings.Replace(server.BaseURL, "http", "ws", 1) + "/ws"

	u, _ := url.Parse(wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer conn.Close()

	// 3. Subscribe to "global_notifications"
	subscribeMsg := map[string]string{
		"type":    "subscribe",
		"channel": "global_notifications",
	}
	err = conn.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Wait for subscription confirmation (info message)
	var infoMsg struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Data    string `json:"data"`
	}
	err = conn.ReadJSON(&infoMsg)
	require.NoError(t, err)
	assert.Equal(t, "info", infoMsg.Type)
	assert.Equal(t, "global_notifications", infoMsg.Channel)

	// 4. Perform Action: Login via REST API (should trigger broadcast)
	f := fixtures.NewUserFactory(server.DB)
	user := f.Create() // Creates a default user

	loginPayload := map[string]any{
		"username": user.Username,
		"password": "password123", // Factory default
	}

	// We use the REST client to login
	resp := server.Client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	// 5. Verify Notification Received via WebSocket
	// The server should broadcast "user_login" event to "global_notifications" channel

	// Set a timeout for reading the message
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	var notification struct {
		Type    string `json:"type"`
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}

	// Note: The message might be wrapped in ServerMessage structure or raw payload
	// depending on how BroadcastToChannel is implemented.
	// WsManager.handleBroadcast sends msg.Message directly to client.Send.
	// AuthUsecase.Login sends a JSON payload.

	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	err = json.Unmarshal(message, &notification)
	require.NoError(t, err)

	assert.Equal(t, "user_login", notification.Type)
	assert.Equal(t, user.ID, notification.UserID)
}
