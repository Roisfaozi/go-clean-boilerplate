package ws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketManager_RedisIntegration(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Use separate clients for each manager to avoid connection closing issues
	rdb1 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdb2 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = rdb1.Close() }()
	defer func() { _ = rdb2.Close() }()

	// Setup Manager 1 (Node 1)
	manager1, server1 := setupTestServer(rdb1)
	defer server1.Close()
	defer manager1.Stop()

	// Setup Manager 2 (Node 2)
	manager2, server2 := setupTestServer(rdb2)
	defer server2.Close()
	defer manager2.Stop()

	// Wait for managers to start and subscribe to redis
	time.Sleep(100 * time.Millisecond)

	// Client 1 connects to Node 1 and subscribes to "global-channel"
	c1, err := connectClient(server1.URL)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }()

	err = c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "global-channel"})
	require.NoError(t, err)
	_, err = waitForMessage(c1, "info", "global-channel")
	require.NoError(t, err)

	// Client 2 connects to Node 2 and subscribes to "global-channel"
	c2, err := connectClient(server2.URL)
	require.NoError(t, err)
	defer func() { _ = c2.Close() }()

	err = c2.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "global-channel"})
	require.NoError(t, err)
	_, err = waitForMessage(c2, "info", "global-channel")
	require.NoError(t, err)

	// Broadcast from Node 1
	msgContent := map[string]string{"msg": "hello from node 1"}
	msgBytes, _ := json.Marshal(msgContent)

	// Retry loop for ensuring c2 receives the message (Redis subscription propagation delay)
	// We need to wait a bit longer for the initial subscription to propagate across the "cluster" (miniredis)
	time.Sleep(500 * time.Millisecond)

	var msg2 *ws.ServerMessage
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		manager1.BroadcastToChannel("global-channel", msgBytes)

		// Check if c2 received it
		_ = c2.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		var receivedMsg ws.ServerMessage
		if err := c2.ReadJSON(&receivedMsg); err == nil {
			if receivedMsg.Type == "message" && receivedMsg.Channel == "global-channel" {
				msg2 = &receivedMsg
				break
			}
		}

		// Backoff slightly
		time.Sleep(100 * time.Millisecond)
	}
	_ = c2.SetReadDeadline(time.Time{})

	// Verify c1 (connected to Node 1) receives it (Local broadcast)
	// Since we might have broadcasted multiple times, just getting one is enough.
	msg1, err := waitForMessage(c1, "message", "global-channel")
	assert.NoError(t, err)
	assert.NotNil(t, msg1)

	// Verify c2 (connected to Node 2) receives it (Redis broadcast)
	require.NotNil(t, msg2, "Failed to receive message on c2 via Redis")

	// Verify content
	// msg2.Data should be interface{} matching msgContent
	dataMap, ok := msg2.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello from node 1", dataMap["msg"])
}

func TestWebSocketManager_Redis_ExternalPublish(t *testing.T) {
	// Test that if an external system publishes to Redis, the manager picks it up
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	manager, server := setupTestServer(rdb)
	defer server.Close()
	defer manager.Stop()

	c1, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }()

	channel := "external-channel"
	err = c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: channel})
	require.NoError(t, err)
	_, err = waitForMessage(c1, "info", channel)
	require.NoError(t, err)

	// Prefix is "test_ws:" defined in setupTestServer
	prefix := "test_ws:"
	redisChannel := prefix + channel

	// Wait for Redis subscription to be active
	// The manager subscribes to "test_ws:*" (pattern), so we check NumPat.
	require.Eventually(t, func() bool {
		// Miniredis supports PubSubNumPat but go-redis API for it is straightforward.
		// However, the manager uses PSubscribe with pattern "prefix*".
		// We can try checking if publishing to a channel in that pattern reaches subscribers?
		// Or just check NumPat.
		// rdb.PubSubNumPat() returns map[string]int64 in some versions or int64?
		// Actually, standard redis command PUBSUB NUMPAT returns count.
		count, err := rdb.PubSubNumPat(context.Background()).Result()
		return err == nil && count > 0
	}, 5*time.Second, 100*time.Millisecond, "Manager failed to subscribe to Redis pattern")

	// The manager expects the payload to be the message bytes (e.g., JSON of ServerMessage or just data)
	// handleBroadcast receives the payload and wraps it in {type: "message", ...} IF it's coming from Redis?
	// No, handleBroadcast wraps msg.Message in "data" field of envelope.

	// Let's send a simple JSON object
	payload := `{"text":"external hello"}`

	// Publish once, since we confirmed subscription is active
	err = rdb.Publish(context.Background(), redisChannel, payload).Err()
	require.NoError(t, err)

	// Verify c1 receives it
	msg, err := waitForMessage(c1, "message", channel)
	require.NoError(t, err)

	// Reset deadline
	_ = c1.SetReadDeadline(time.Time{})

	require.NotNil(t, msg, "Failed to receive message from Redis subscription after retries")

	// The data field of the message should contain the payload parsed as JSON if possible
	// The manager does: "data": json.RawMessage(msg.Message)
	// So msg.Data (in ServerMessage struct) will be unmarshaled payload.
	dataMap, ok := msg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "external hello", dataMap["text"])
}
