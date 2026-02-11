package sse_test

import (
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSE_ClientCount(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	assert.Equal(t, 0, manager.ClientCount())

	c1 := &sse.Client{Channel: make(chan sse.Event, 1)}
	c2 := &sse.Client{Channel: make(chan sse.Event, 1)}

	manager.RegisterClient(c1)
	time.Sleep(50 * time.Millisecond) // wait for goroutine processing
	assert.Equal(t, 1, manager.ClientCount())

	manager.RegisterClient(c2)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 2, manager.ClientCount())

	manager.UnregisterClient(c1)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, manager.ClientCount())

	manager.UnregisterClient(c2)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, manager.ClientCount())
}

func TestSSE_BroadcastToMultipleClients(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	c1Chan := make(chan sse.Event, 1)
	c2Chan := make(chan sse.Event, 1)
	c1 := &sse.Client{Channel: c1Chan}
	c2 := &sse.Client{Channel: c2Chan}

	manager.RegisterClient(c1)
	manager.RegisterClient(c2)
	time.Sleep(50 * time.Millisecond)

	manager.Broadcast("multi-event", map[string]string{"key": "value"})

	// Both clients should receive the event
	select {
	case event := <-c1Chan:
		assert.Equal(t, "multi-event", event.Name)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client 1 did not receive broadcast")
	}

	select {
	case event := <-c2Chan:
		assert.Equal(t, "multi-event", event.Name)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client 2 did not receive broadcast")
	}
}

func TestSSE_SlowClientEviction(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	// Use unbuffered channel to simulate a slow client
	slowChan := make(chan sse.Event) // unbuffered - will block
	slowClient := &sse.Client{Channel: slowChan}

	fastChan := make(chan sse.Event, 10) // buffered - will not block
	fastClient := &sse.Client{Channel: fastChan}

	manager.RegisterClient(slowClient)
	manager.RegisterClient(fastClient)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, manager.ClientCount())

	// Broadcast - slow client's channel is full/blocking, should be evicted
	manager.Broadcast("eviction-test", "data")
	time.Sleep(100 * time.Millisecond)

	// Fast client should get the message
	select {
	case event := <-fastChan:
		assert.Equal(t, "eviction-test", event.Name)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Fast client should have received the event")
	}

	// Slow client should be evicted
	assert.Equal(t, 1, manager.ClientCount())
}

func TestSSE_UnregisterNonExistentClient(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	nonExistent := &sse.Client{Channel: make(chan sse.Event, 1)}

	// Should not panic or error
	manager.UnregisterClient(nonExistent)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, manager.ClientCount())
}

func TestSSE_SetLogger(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	customLogger := logrus.New()
	customLogger.SetLevel(logrus.DebugLevel)

	// Should not panic
	manager.SetLogger(customLogger)
}

func TestSSE_BroadcastWithComplexData(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event, 1)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)

	complexData := map[string]interface{}{
		"user_id":   "u123",
		"action":    "login",
		"timestamp": 1234567890,
		"nested":    map[string]string{"key": "val"},
		"list":      []int{1, 2, 3},
	}

	manager.Broadcast("complex-event", complexData)

	select {
	case event := <-clientChan:
		assert.Equal(t, "complex-event", event.Name)
		dataMap, ok := event.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "u123", dataMap["user_id"])
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client did not receive complex broadcast")
	}
}

func TestSSE_BroadcastEmptyEventName(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event, 1)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)
	time.Sleep(50 * time.Millisecond)

	manager.Broadcast("", "empty-name-data")

	select {
	case event := <-clientChan:
		assert.Equal(t, "", event.Name)
		assert.Equal(t, "empty-name-data", event.Data)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Client did not receive event with empty name")
	}
}
