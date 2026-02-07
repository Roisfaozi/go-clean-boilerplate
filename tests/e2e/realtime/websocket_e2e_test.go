//go:build e2e
// +build e2e

package realtime

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketE2E_NotificationFlow(t *testing.T) {

	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	registerPayload := map[string]any{
		"name":     "WS Test User",
		"username": "wstestuser_" + timestamp(),
		"email":    "wstest_" + timestamp() + "@example.com",
		"password": "password123",
	}

	w := server.Client.POST("/api/v1/auth/register", registerPayload)
	require.Equal(t, 201, w.StatusCode)

	var registerResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
			User        struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.BodyBytes, &registerResp)
	require.NoError(t, err)
	accessToken := registerResp.Data.AccessToken
	userID := registerResp.Data.User.ID

	// 2. Get User Organization
	wOrg := server.Client.GET("/api/v1/organizations/me", setup.WithAuth(accessToken))
	require.Equal(t, 200, wOrg.StatusCode)

	var orgResp struct {
		Data struct {
			Organizations []struct {
				ID string `json:"id"`
			} `json:"organizations"`
		} `json:"data"`
	}
	err = json.Unmarshal(wOrg.BodyBytes, &orgResp)
	require.NoError(t, err)
	require.NotEmpty(t, orgResp.Data.Organizations, "User should have at least one organization")
	orgID := orgResp.Data.Organizations[0].ID

	// 3. Connect to WebSocket
	wsURL := strings.Replace(server.BaseURL, "http", "ws", 1) + "/ws"
	u, _ := url.Parse(wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer conn.Close()

	// 4. Subscribe to Organization Channel
	subscribeMsg := map[string]string{
		"type":    "subscribe",
		"channel": "org_" + orgID + "_notifications",
	}
	err = conn.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// Verify subscription info message
	var infoMsg struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Data    string `json:"data"`
	}
	err = conn.ReadJSON(&infoMsg)
	require.NoError(t, err)
	assert.Equal(t, "info", infoMsg.Type)
	assert.Equal(t, "org_"+orgID+"_notifications", infoMsg.Channel)

	// 5. Trigger Notification (Login)
	t.Log("Waiting for subscription to propagate...")
	time.Sleep(1 * time.Second)
	t.Log("Triggering Login to generate notification...")

	// We login again to trigger the "user_login" event
	loginPayload := map[string]any{
		"username": registerPayload["username"],
		"password": registerPayload["password"],
	}

	wLogin := server.Client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, wLogin.StatusCode)
	t.Log("Login successful")

	// 6. Verify Notification
	t.Log("Waiting for notification...")
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	var wsWrapper struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		Data    struct {
			Type    string `json:"type"`
			UserID  string `json:"user_id"`
			Message string `json:"message"`
		} `json:"data"`
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Logf("ReadMessage failed: %v", err)
	}
	require.NoError(t, err)
	t.Logf("Received message: %s", string(message))

	err = json.Unmarshal(message, &wsWrapper)
	require.NoError(t, err)

	assert.Equal(t, "message", wsWrapper.Type)
	assert.Equal(t, "user_login", wsWrapper.Data.Type)
	assert.Equal(t, userID, wsWrapper.Data.UserID)
}

func timestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func TestPresenceE2E_IsolationAndEvents(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	createUser := func(namePrefix string) (string, string, string) { // returns accessToken, userID, orgID
		registerPayload := map[string]any{
			"name":     namePrefix,
			"username": namePrefix + "_" + timestamp(),
			"email":    namePrefix + "_" + timestamp() + "@example.com",
			"password": "password123",
		}
		w := server.Client.POST("/api/v1/auth/register", registerPayload)
		require.Equal(t, 201, w.StatusCode)

		var resp struct {
			Data struct {
				AccessToken string `json:"access_token"`
				User        struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"data"`
		}
		json.Unmarshal(w.BodyBytes, &resp)
		token := resp.Data.AccessToken
		uid := resp.Data.User.ID

		// Create a unique organization
		orgPayload := map[string]any{
			"name": namePrefix + " Org " + timestamp(),
			"slug": strings.ToLower(namePrefix) + "-org-" + timestamp(),
		}
		wCreateOrg := server.Client.POST("/api/v1/organizations", orgPayload, setup.WithAuth(token))
		require.Equal(t, 201, wCreateOrg.StatusCode)

		var orgCreateResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		json.Unmarshal(wCreateOrg.BodyBytes, &orgCreateResp)
		orgID := orgCreateResp.Data.ID

		return token, uid, orgID
	}

	connectWS := func(token, orgID string) *websocket.Conn {
		wsURL := strings.Replace(server.BaseURL, "http", "ws", 1) + "/ws?token=" + token
		if orgID != "" {
			wsURL += "&org_id=" + orgID
		}
		u, _ := url.Parse(wsURL)
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		require.NoError(t, err)
		return conn
	}

	tokenA, uidA, org1ID := createUser("UserA")
	connA := connectWS(tokenA, org1ID)
	defer connA.Close()

	channelOrg1 := "presence:org:" + org1ID
	connA.WriteJSON(map[string]string{
		"type":    "subscribe",
		"channel": channelOrg1,
	})
	_, _, _ = connA.ReadMessage()

	// Wait for async registration
	time.Sleep(500 * time.Millisecond)

	wPresence := server.Client.GET("/api/v1/organizations/"+org1ID+"/presence", setup.WithAuth(tokenA))
	require.Equal(t, 200, wPresence.StatusCode)
	assert.Contains(t, string(wPresence.BodyBytes), uidA)

	connA2 := connectWS(tokenA, org1ID)
	defer connA2.Close()
	connA2.WriteJSON(map[string]string{"type": "subscribe", "channel": channelOrg1})

	connA.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgA, err := connA.ReadMessage()
	require.NoError(t, err)
	t.Logf("A received: %s", string(msgA))

	assert.Contains(t, string(msgA), "\"event\":\"join\"")
	assert.Contains(t, string(msgA), uidA)

	tokenC, _, org2ID := createUser("UserC")
	connC := connectWS(tokenC, org2ID)
	defer connC.Close()
	channelOrg2 := "presence:org:" + org2ID
	connC.WriteJSON(map[string]string{"type": "subscribe", "channel": channelOrg2})
	// Trigger event in Org 1 again (A3 connects)
	connA3 := connectWS(tokenA, org1ID)
	defer connA3.Close()
	connA3.WriteJSON(map[string]string{"type": "subscribe", "channel": channelOrg1})

	// A should receive A3 join event (drain buffer)
	connA.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgA3Join, err := connA.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msgA3Join), "\"event\":\"join\"")

	// C should NOT receive Org 1 events
	// C might receive its own join event, so we must check the channel if a message arrives
	connC.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, msgC, errC := connC.ReadMessage()
	if errC == nil {
		// If message received, it MUST NOT be from channelOrg1
		assert.NotContains(t, string(msgC), channelOrg1, "User C received message from Org 1 channel")
		// If it is from channelOrg2, that is acceptable (own join event)
	}

	connA3.Close()

	connA.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, msgLeave, err := connA.ReadMessage()
	require.NoError(t, err)
	t.Logf("A received leave: %s", string(msgLeave))
	assert.Contains(t, string(msgLeave), "\"event\":\"leave\"")
}
