package sse_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	manager := sse.NewManager()
	assert.NotNil(t, manager)
	manager.Stop()
}

func TestManager_RegisterAndUnregister(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event)
	client := &sse.Client{Channel: clientChan}

	manager.RegisterClient(client)

	// Wait for registration to process
	for i := 0; i < 10; i++ {
		if manager.ClientCount() == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 1, manager.ClientCount())

	manager.UnregisterClient(client)

	for i := 0; i < 10; i++ {
		if manager.ClientCount() == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 0, manager.ClientCount())
}

func TestManager_Broadcast(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event, 1)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)

	time.Sleep(50 * time.Millisecond)

	eventName := "test-event"
	eventData := "hello"
	manager.Broadcast(eventName, eventData)

	select {
	case event := <-clientChan:
		assert.Equal(t, eventName, event.Name)
		assert.Equal(t, eventData, event.Data)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client did not receive broadcast message")
	}
}

func TestManager_SetLogger(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	logger := logrus.New()
	manager.SetLogger(logger)
	// Just verify no panic
	assert.NotNil(t, manager)
}

func TestManager_ServeHTTP(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/events", manager.ServeHTTP())

	server := httptest.NewServer(r)
	defer server.Close()

	// Connect to the server
	client := &http.Client{
		Timeout: 2 * time.Second, // Prevent hanging test
	}
	req, _ := http.NewRequest("GET", server.URL+"/events", nil)

	// Create a context with cancellation to simulate client disconnect
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req = req.WithContext(ctx)

	// We start the request in a goroutine because it might block depending on how httptest/client behaves
	// But actually client.Do returns as soon as headers are received for streaming?
	// Standard http client waits for headers.

	resp, err := client.Do(req)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	// Don't defer body close immediately, we want to read first

	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Wait for registration
	for i := 0; i < 20; i++ {
		if manager.ClientCount() == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 1, manager.ClientCount())

	// Broadcast
	manager.Broadcast("ping", "pong")

	// Read from body
	buf := make([]byte, 1024)
	// Read should return at least some bytes when data arrives
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	data := string(buf[:n])
	assert.Contains(t, data, "event: ping")
	assert.Contains(t, data, "data: \"pong\"")

	// Close response body to signal disconnect?
	// Or cancel context.
	_ = resp.Body.Close()
	cancel()

	// Verify cleanup
	for i := 0; i < 20; i++ {
		if manager.ClientCount() == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, 0, manager.ClientCount())
}

func TestManager_SlowClient(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	// Create a client with unbuffered channel
	clientChan := make(chan sse.Event)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Broadcast should not block even if client is not reading.
	// The manager implementation uses a select with default to handle slow clients.
	// Since channel is unbuffered and we don't read from it, the send will block,
	// triggering the default case which removes the client.
	manager.Broadcast("test", "data")

	// Wait for manager loop to process
	for i := 0; i < 10; i++ {
		if manager.ClientCount() == 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Client should be removed
	assert.Equal(t, 0, manager.ClientCount())
}
