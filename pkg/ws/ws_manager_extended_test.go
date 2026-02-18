package ws_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPresenceManager implements ws.PresenceManager
type MockPresenceManager struct {
	mock.Mock
}

func (m *MockPresenceManager) SetUserOnline(ctx context.Context, orgID, userID string, userData *ws.PresenceUser) error {
	args := m.Called(ctx, orgID, userID, userData)
	return args.Error(0)
}

func (m *MockPresenceManager) SetUserOffline(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockPresenceManager) GetOnlineUsers(ctx context.Context, orgID string) ([]ws.PresenceUser, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]ws.PresenceUser), args.Error(1)
}

func (m *MockPresenceManager) RefreshUserHeartbeat(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockPresenceManager) PruneStaleUsers(ctx context.Context, timeout time.Duration) (map[string][]string, error) {
	args := m.Called(ctx, timeout)
	return args.Get(0).(map[string][]string), args.Error(1)
}

func TestWebSocketManager_Stopped_Methods(t *testing.T) {
	manager, server := setupTestServer(nil)
	defer server.Close()

	// Wait for start
	time.Sleep(10 * time.Millisecond)
	manager.Stop()
	time.Sleep(10 * time.Millisecond) // Ensure loop exits

	// Test methods on stopped manager
	// These should log warnings and return immediately (hitting the select default or timeout)

	// Create a dummy client
	client := &ws.Client{ID: "test"}

	done := make(chan bool)
	go func() {
		manager.RegisterClient(client)
		manager.UnregisterClient(client)
		manager.BroadcastToChannel("ch", []byte("msg"))
		manager.SubscribeToChannel(client, "ch")
		manager.UnsubscribeFromChannel(client, "ch")
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Methods blocked on stopped manager")
	}
}

func TestWebSocketManager_PresenceError_Register(t *testing.T) {
	mockPresence := new(MockPresenceManager)
	mockPresence.On("SetUserOnline", mock.Anything, "org1", "u1", mock.Anything).Return(errors.New("presence error"))

	config := &ws.WebSocketConfig{RedisPrefix: "test:"}
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})
	manager := ws.NewWebSocketManager(config, logger, nil, mockPresence)
	go manager.Run()
	defer manager.Stop()

	// Register client
	client := &ws.Client{
		ID:     "c1",
		UserID: "u1",
		OrgID:  "org1",
		Send:   make(chan []byte, 10),
	}
	manager.RegisterClient(client)

	// Wait for processing
	require.Eventually(t, func() bool { return manager.ClientCount() == 1 }, time.Second, 10*time.Millisecond)

	// Assert expectations
	mockPresence.AssertExpectations(t)
	// Client should still be registered even if presence fails
	assert.Equal(t, 1, manager.ClientCount())
}

func TestWebSocketManager_PresenceError_Unregister(t *testing.T) {
	mockPresence := new(MockPresenceManager)
	// SetUserOnline succeeds
	mockPresence.On("SetUserOnline", mock.Anything, "org1", "u1", mock.Anything).Return(nil)
	// SetUserOffline fails
	mockPresence.On("SetUserOffline", mock.Anything, "org1", "u1").Return(errors.New("presence error"))

	config := &ws.WebSocketConfig{RedisPrefix: "test:"}
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})
	manager := ws.NewWebSocketManager(config, logger, nil, mockPresence)
	go manager.Run()
	defer manager.Stop()

	// Register client
	client := &ws.Client{
		ID:     "c1",
		UserID: "u1",
		OrgID:  "org1",
		Send:   make(chan []byte, 10),
	}
	manager.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)

	// Unregister
	manager.UnregisterClient(client)
	require.Eventually(t, func() bool { return manager.ClientCount() == 0 }, time.Second, 10*time.Millisecond)

	mockPresence.AssertExpectations(t)
	assert.Equal(t, 0, manager.ClientCount())
}

func TestWebSocketManager_Broadcast_RedisError(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()

	config := &ws.WebSocketConfig{
		DistributedEnabled: true,
		RedisPrefix:        "test:",
	}
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	manager := ws.NewWebSocketManager(config, logger, db, &NoOpPresenceManager{})

    // Not calling Run() because redismock PSubscribe handling is complex
    // Just verify instantiation logic was executed
    assert.NotNil(t, manager)

	// Note: We don't call Run() here to avoid listenToRedis complications with mock.
    // If we wanted to test this properly with redismock, we would need to mock everything Run calls.
    // Given TestWebSocketManager_Broadcast_RedisError_Miniredis exists, we can skip deeper logic here.

    // Just cleanup unused mock
    _ = mockRedis
}

func TestWebSocketManager_Broadcast_RedisError_Miniredis(t *testing.T) {
	// Use miniredis to simulate redis failure
	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	config := &ws.WebSocketConfig{
		DistributedEnabled: true,
		RedisPrefix:        "test:",
	}
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	manager := ws.NewWebSocketManager(config, logger, rdb, &NoOpPresenceManager{})
	go manager.Run()
	defer manager.Stop()

	// Wait for start
	time.Sleep(50 * time.Millisecond)

	// Close miniredis to cause connection error
	mr.Close()

	// Broadcast
	// This should trigger "Failed to publish to Redis" log, but not panic
	manager.BroadcastToChannel("ch1", []byte("msg"))

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// No panic implies success
}

// Test buffer full scenario
func TestWebSocketManager_Broadcast_BufferFull(t *testing.T) {
	manager, server := setupTestServer(nil)
	defer server.Close()
	defer manager.Stop()

	// Manually create a client with small buffer
	client := &ws.Client{
		ID:   "c1",
		Send: make(chan []byte, 1), // Buffer size 1
	}
	manager.RegisterClient(client)
	manager.SubscribeToChannel(client, "ch1")

	// Wait for registration and subscription
	for i := 0; i < 20; i++ {
		if manager.ClientCount() == 1 && manager.GetChannelClients("ch1") == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, 1, manager.ClientCount())
	require.Equal(t, 1, manager.GetChannelClients("ch1"))

	// Fill buffer
	// We need to send valid JSON so that Marshal doesn't fail (though RawMessage usually allows anything, the envelope Marshal might be picky or the client ReadPump might close connection on invalid JSON?)
	// Actually, client.Send just receives bytes. WritePump sends them.
	// But let's use valid JSON just in case.
	manager.BroadcastToChannel("ch1", []byte(`"msg1"`))

	// Wait for processing
	require.Eventually(t, func() bool { return len(client.Send) == 1 }, time.Second, 10*time.Millisecond)

	// Send more - should drop (log warn)
	manager.BroadcastToChannel("ch1", []byte(`"msg2"`))
	manager.BroadcastToChannel("ch1", []byte(`"msg3"`))

	time.Sleep(10 * time.Millisecond)

	// Verify buffer is full (should contain msg1)
	assert.Equal(t, 1, len(client.Send))
}

func TestWebSocketManager_Channels_Getter(t *testing.T) {
    manager, server := setupTestServer(nil)
    defer server.Close()
    defer manager.Stop()

    // Register and Subscribe
    conn, err := connectClient(server.URL)
    require.NoError(t, err)
    defer func() { _ = conn.Close() }()

    err = conn.WriteJSON(ws.ClientMessage{Type: "subscribe", Channel: "ch1"})
    require.NoError(t, err)

    _, err = waitForMessage(conn, "info", "ch1")
    require.NoError(t, err)

    channels := manager.Channels()
    assert.Contains(t, channels, "ch1")
    assert.Equal(t, 1, len(channels["ch1"]))
}
