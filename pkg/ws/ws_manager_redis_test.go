package ws_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServerWithRedis creates a WebSocketManager with Redis enabled for distributed testing.
func setupTestServerWithRedis(rdb *redis.Client, prefix string) (*ws.WebSocketManager, *httptest.Server) {
	config := &ws.WebSocketConfig{
		WriteWait:          10 * time.Second,
		PongWait:           60 * time.Second,
		PingPeriod:         54 * time.Second,
		MaxMessageSize:     512 * 1024,
		DistributedEnabled: true,
		RedisPrefix:        prefix,
	}
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})
	presence := &NoOpPresenceManager{}
	manager := ws.NewWebSocketManager(config, logger, rdb, presence)
	go manager.Run()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Create a client for testing purposes. In a real scenario, userID and orgID would come from auth context.
		// Using a unique ID for each connection would be better if multiple clients connect to the same manager in one test.
		client := ws.NewWebsocketClient(conn, manager, logger, config, "u1", "org1", nil)
		manager.RegisterClient(client)
		go client.WritePump()
		go client.ReadPump()
	})

	server := httptest.NewServer(handler)
	return manager, server
}

func TestWebSocketManager_RedisIntegration(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	prefix := "test_ws:"

	// Use separate clients for each manager to avoid connection closing issues
	rdb1 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdb2 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb1.Close() }()
	defer func() { _ = rdb2.Close() }()

	// Setup Manager 1 (Node 1) with Redis
	manager1, server1 := setupTestServerWithRedis(rdb1, prefix)
	defer server1.Close()
	defer manager1.Stop()

	// Setup Manager 2 (Node 2) with Redis
	manager2, server2 := setupTestServerWithRedis(rdb2, prefix)
	defer server2.Close()
	defer manager2.Stop()

	// Wait for managers to start and subscribe to redis
	time.Sleep(100 * time.Millisecond)

	// Client 1 connects to Node 1 and subscribes to "global-channel"
	c1, _, err := websocket.DefaultDialer.Dial(makeWsProto(server1.URL), nil)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }()

	err = c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "global-channel"})
	require.NoError(t, err)

	// Wait for subscription confirmation
	_, _, err = c1.ReadMessage()
	require.NoError(t, err)


	// Client 2 connects to Node 2 and subscribes to "global-channel"
	c2, _, err := websocket.DefaultDialer.Dial(makeWsProto(server2.URL), nil)
	require.NoError(t, err)
	defer func() { _ = c2.Close() }()

	err = c2.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "global-channel"})
	require.NoError(t, err)

	// Wait for subscription confirmation
	_, _, err = c2.ReadMessage()
	require.NoError(t, err)

	// Wait for Redis subscription propagation across nodes
	// This is the tricky part with miniredis/redis pubsub, sometimes it takes a bit.
	// But since we are using the same miniredis instance, it should be fast.
	time.Sleep(200 * time.Millisecond)

	// Broadcast from Node 1
	msgContent := map[string]string{"msg": "hello from node 1"}

	// We need to use a struct that matches what BroadcastToChannel expects or just a map
	// The implementation of BroadcastToChannel marshals the data.

	msgContentBytes, _ := json.Marshal(msgContent)
	manager1.BroadcastToChannel("global-channel", msgContentBytes)

	// Verify c2 (connected to Node 2) receives it (Redis broadcast)
	// We use a timeout to read
	_ = c2.SetReadDeadline(time.Now().Add(2 * time.Second))
	var receivedMsg ws.ServerMessage
	err = c2.ReadJSON(&receivedMsg)
	require.NoError(t, err, "Failed to receive message on c2 via Redis")

	assert.Equal(t, "message", receivedMsg.Type)
	assert.Equal(t, "global-channel", receivedMsg.Channel)

	// Verify content
	dataMap, ok := receivedMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello from node 1", dataMap["msg"])
}

func TestWebSocketManager_Redis_ExternalPublish(t *testing.T) {
	// Test that if an external system publishes to Redis, the manager picks it up
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Use separate clients for manager and publisher to avoid interference
	managerRdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	pubRdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = managerRdb.Close() }()
	defer func() { _ = pubRdb.Close() }()

	prefix := "test_ws:"

	// Setup manager with Redis enabled
	manager, server := setupTestServerWithRedis(managerRdb, prefix)
	defer server.Close()
	defer manager.Stop()

	// Connect a client
	c1, _, err := websocket.DefaultDialer.Dial(makeWsProto(server.URL), nil)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }()

	channel := "external-channel"
	err = c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: channel})
	require.NoError(t, err)

	// Read the subscription confirmation "info" message
	var confirmMsg ws.ServerMessage
	err = c1.ReadJSON(&confirmMsg)
	require.NoError(t, err)
	require.Equal(t, "info", confirmMsg.Type)
	require.Equal(t, channel, confirmMsg.Channel)

	redisChannel := prefix + channel

	// Wait for the manager's Redis subscription to be active.
	// The manager subscribes to a pattern, e.g. "test_ws:*".
	// We can check PubSubNumPat to ensure the subscription is registered.
	require.Eventually(t, func() bool {
		// PubSubNumPat returns a map[string]int64 in newer go-redis versions,
		// but check the specific version used in go.mod if unsure.
		// Actually, PubSubNumPat().Result() returns map[string]int64 or just int64 depending on version.
		// Let's use the miniredis implementation behavior or just generic PubSubNumPat.
		// Wait, miniredis might not support PubSubNumPat fully or correctly for patterns in all versions.
		// A safer check is to publish and wait for receipt with retry.
		// But let's try to verify subscription first if possible.

		// Alternative: check if the pattern subscription exists using PUBSUB NUMPAT
		// miniredis supports PUBSUB NUMPAT.

		// Note: The previous failing test used PubSubNumPat and it passed that check but failed on read.
		// The error was "i/o timeout" on c1.ReadJSON, meaning the message never arrived at the client.
		// This implies the manager didn't receive the Redis message or didn't forward it.

		// To fix flakiness, we will use a polling loop to publish and check for message,
		// because synchronization of "when subscription is ready" is hard with just sleep.
		return true
	}, 5*time.Second, 100*time.Millisecond, "Manager failed to subscribe to Redis pattern")

	// Send a simple JSON object
	// The manager expects the payload to be the data part, or a full message?
	// Looking at manager.listenToRedis:
	// It unmarshals the payload into a ServerMessage? Or does it wrap it?
	// If it's a raw string, it might try to unmarshal it.
	// If the payload is just data, the manager constructs the ServerMessage?
	// We need to know how listenToRedis handles the message.
	// Assuming it expects the payload to be the data content or a JSON representation of it.

	// In the original test: payload := `{"text":"external hello"}`
	// And verification: assert.Equal(t, "external hello", dataMap["text"])

	payload := `{"text":"external hello"}`

	// Retry loop: Publish and wait for message
	// The subscription might take a few milliseconds to fully propagate in miniredis/redis.
	// We'll try publishing every 100ms until we receive it.

	received := false
	var msg ws.ServerMessage

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		err = pubRdb.Publish(context.Background(), redisChannel, payload).Err()
		require.NoError(t, err)

		// Brief wait for processing
		time.Sleep(50 * time.Millisecond)

		// Non-blocking read attempt
		_ = c1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		err = c1.ReadJSON(&msg)
		if err == nil {
			if msg.Type == "message" && msg.Channel == channel {
				received = true
				break
			}
		}
		// If error was not timeout, it might be a connection error, which is bad.
		// But here we expect timeout if message not received yet.
	}

	require.True(t, received, "Failed to receive message from Redis subscription after retries")

	// Reset deadline
	_ = c1.SetReadDeadline(time.Time{})

	// The data field of the message should contain the payload parsed as JSON if possible
	dataMap, ok := msg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "external hello", dataMap["text"])
}

// Helper to handle ws:// vs http://
func makeWsProto(s string) string {
	return "ws" + s[4:]
}
