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

	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	wsURL := strings.Replace(server.BaseURL, "http", "ws", 1) + "/ws"

	u, _ := url.Parse(wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer conn.Close()

	subscribeMsg := map[string]string{
		"type":    "subscribe",
		"channel": "global_notifications",
	}
	err = conn.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	var infoMsg struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Data    string `json:"data"`
	}
	err = conn.ReadJSON(&infoMsg)
	require.NoError(t, err)
	assert.Equal(t, "info", infoMsg.Type)
	assert.Equal(t, "global_notifications", infoMsg.Channel)

	f := fixtures.NewUserFactory(server.DB)
	user := f.Create()

	loginPayload := map[string]any{
		"username": user.Username,
		"password": "password123",
	}

	resp := server.Client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// The WebSocket Manager wraps messages in an envelope:
	// { "type": "message", "channel": "...", "data": { ... } }
	var envelope struct {
		Type    string          `json:"type"`
		Channel string          `json:"channel"`
		Data    json.RawMessage `json:"data"`
	}

	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	err = json.Unmarshal(message, &envelope)
	require.NoError(t, err)

	// Verify Envelope
	assert.Equal(t, "message", envelope.Type)
	assert.Equal(t, "global_notifications", envelope.Channel)

	// Verify Inner Payload
	var notification struct {
		Type    string `json:"type"`
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(envelope.Data, &notification)
	require.NoError(t, err)

	assert.Equal(t, "user_login", notification.Type)
	assert.Equal(t, user.ID, notification.UserID)
}
