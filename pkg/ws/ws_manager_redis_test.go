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

	// Helper to create redis client with optimal settings for miniredis
	newRedisClient := func(addr string) *redis.Client {
		return redis.NewClient(&redis.Options{
			Addr:         addr,
			MinIdleConns: 2, // Need at least 2 connections (one for PubSub, one for Publish)
			PoolSize:     10,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		})
	}

	// Setup Manager 1 (Node 1)
	rdb1 := newRedisClient(mr.Addr())
	defer func() { _ = rdb1.Close() }()
	manager1, server1 := setupTestServer(rdb1)
	defer server1.Close()
	defer manager1.Stop()

	// Setup Manager 2 (Node 2)
	rdb2 := newRedisClient(mr.Addr())
	defer func() { _ = rdb2.Close() }()
	manager2, server2 := setupTestServer(rdb2)
	defer server2.Close()
	defer manager2.Stop()

	// Wait for managers to start and subscribe to redis
	time.Sleep(3 * time.Second)

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

	// Determine expected payload on Redis
	// Manager publishes raw message bytes to redis channel.
	// Then other managers receive it and wrap in envelope.

	// Wait, Manager.BroadcastToChannel(channel, message)
	// It publishes message to redis.
	// It also broadcasts locally.

	manager1.BroadcastToChannel("global-channel", msgBytes)

	// Verify c1 (connected to Node 1) receives it (Local broadcast)
	msg1, err := waitForMessage(c1, "message", "global-channel")
	require.NoError(t, err)
	require.NotNil(t, msg1)

	// Verify c2 (connected to Node 2) receives it (Redis broadcast)
	msg2, err := waitForMessage(c2, "message", "global-channel")
	require.NoError(t, err)
	require.NotNil(t, msg2)

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
		Addr:         mr.Addr(),
		MinIdleConns: 1,
		PoolSize:     1,
	})
	defer func() { _ = rdb.Close() }()

	manager, server := setupTestServer(rdb)
	defer server.Close()
	defer manager.Stop()

	// Wait for subscription
	time.Sleep(2 * time.Second)

	c1, err := connectClient(server.URL)
	require.NoError(t, err)
	defer func() { _ = c1.Close() }()

	channel := "external-channel"
	err = c1.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: channel})
	require.NoError(t, err)
	_, err = waitForMessage(c1, "info", channel)
	require.NoError(t, err)

	// Publish directly to Redis
	// Prefix is "test_ws:" defined in setupTestServer
	prefix := "test_ws:"

	// The manager expects the payload to be the message bytes (e.g., JSON of ServerMessage or just data)
	// handleBroadcast receives the payload and wraps it in {type: "message", ...} IF it's coming from Redis?
	// No, handleBroadcast wraps msg.Message in "data" field of envelope.

	// Let's send a simple JSON object
	payload := `{"text":"external hello"}`
	err = rdb.Publish(context.Background(), prefix+channel, payload).Err()
	require.NoError(t, err)

	// Verify c1 receives it
	msg, err := waitForMessage(c1, "message", channel)
	require.NoError(t, err)

	// The data field of the message should contain the payload parsed as JSON if possible
	// The manager does: "data": json.RawMessage(msg.Message)
	// So msg.Data (in ServerMessage struct) will be unmarshaled payload.
	dataMap, ok := msg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "external hello", dataMap["text"])
}
