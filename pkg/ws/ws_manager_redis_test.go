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

	// Wait for Redis subscription propagation
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

	// Verify c1 (connected to Node 1) receives it via Redis echo
	msg1, err := waitForMessage(c1, "message", "global-channel")
	assert.NoError(t, err)
	assert.NotNil(t, msg1)

	// Verify c2 (connected to Node 2) receives it (Redis broadcast)
	require.NotNil(t, msg2, "Failed to receive message on c2 via Redis")

	// Verify content
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

	prefix := "test_ws:"

	// Setup manager with Redis enabled
	manager, server := setupTestServerWithRedis(rdb, prefix)
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

	redisChannel := prefix + channel

	// Wait for Redis subscription to be active
	// Use a separate client to check and publish
	pubRdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = pubRdb.Close() }()

	require.Eventually(t, func() bool {
		count, err := pubRdb.PubSubNumPat(context.Background()).Result()
		return err == nil && count > 0
	}, 5*time.Second, 100*time.Millisecond, "Manager failed to subscribe to Redis pattern")

	// Send a simple JSON object
	payload := `{"text":"external hello"}`

	// Publish once, since we confirmed subscription is active
	err = pubRdb.Publish(context.Background(), redisChannel, payload).Err()
	require.NoError(t, err)

	// Verify c1 receives it
	msg, err := waitForMessage(c1, "message", channel)
	require.NoError(t, err)

	// Reset deadline
	_ = c1.SetReadDeadline(time.Time{})

	require.NotNil(t, msg, "Failed to receive message from Redis subscription after retries")

	// The data field of the message should contain the payload parsed as JSON if possible
	dataMap, ok := msg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "external hello", dataMap["text"])
}
