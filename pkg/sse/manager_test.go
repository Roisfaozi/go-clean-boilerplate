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

func TestSetLogger(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	manager.SetLogger(logger)
}

func TestClientCount(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	assert.Equal(t, 0, manager.ClientCount())

	clientChan := make(chan sse.Event)
	client := &sse.Client{Channel: clientChan}

	manager.RegisterClient(client)
	// Give time for the goroutine to process registration
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, manager.ClientCount())

	manager.UnregisterClient(client)
	// Give time for the goroutine to process unregistration
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, manager.ClientCount())
}

func TestBroadcast(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event, 1)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)

	time.Sleep(20 * time.Millisecond)

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

func TestBroadcast_BufferFull(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	// Unbuffered channel - write blocks immediately if no reader
	clientChan := make(chan sse.Event)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)
	time.Sleep(20 * time.Millisecond)

	// Broadcast without reading from clientChan
	manager.Broadcast("test", "data")

	// Wait for processing
	time.Sleep(20 * time.Millisecond)

	// Since channel was blocking, the default case should have triggered
	// Client should be removed
	assert.Equal(t, 0, manager.ClientCount())
}

func TestServeHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := sse.NewManager()
	defer manager.Stop()

	r := gin.New()
	r.GET("/events", manager.ServeHTTP())

	server := httptest.NewServer(r)
	defer server.Close()

	respChan := make(chan *http.Response)
	errChan := make(chan error)

	t.Log("Connecting client")
	go func() {
		// Use context with timeout for client
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		respChan <- resp
	}()

	t.Log("Waiting for client registration")
	// Wait for client to be registered
	assert.Eventually(t, func() bool {
		return manager.ClientCount() == 1
	}, 500*time.Millisecond, 10*time.Millisecond)

	t.Log("Broadcasting")
	// Broadcast an event
	manager.Broadcast("test", "data")
	t.Log("Broadcast done")

	var resp *http.Response
	select {
	case resp = <-respChan:
		// ok
	case err := <-errChan:
		t.Fatalf("Client connection failed: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Client connection timed out waiting for headers")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read output
	t.Log("Reading body")
	buf := make([]byte, 1024)
	n, err := resp.Body.Read(buf)
	t.Logf("Read done: n=%d, err=%v", n, err)
	if n == 0 && err == nil {
		// retry once
		time.Sleep(10 * time.Millisecond)
		n, _ = resp.Body.Read(buf)
	}

	// We might get partial reads or error if closed, but we expect data
	output := string(buf[:n])
	assert.Contains(t, output, "event: test")
	assert.Contains(t, output, "data: \"data\"")

	t.Log("Closing body")
	// Close connection to trigger unregister
	_ = resp.Body.Close()

	t.Log("Waiting for unregister")
	assert.Eventually(t, func() bool {
		return manager.ClientCount() == 0
	}, 500*time.Millisecond, 10*time.Millisecond)
}

func TestServeHTTP_JsonMarshalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := sse.NewManager()
	defer manager.Stop()

	r := gin.New()
	r.GET("/events", manager.ServeHTTP())

	server := httptest.NewServer(r)
	defer server.Close()

	respChan := make(chan *http.Response)
	errChan := make(chan error)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		respChan <- resp
	}()

	assert.Eventually(t, func() bool {
		return manager.ClientCount() == 1
	}, 500*time.Millisecond, 10*time.Millisecond)

	// Broadcast unmarshallable data
	manager.Broadcast("error-event", func() {})

	var resp *http.Response
	select {
	case resp = <-respChan:
		// ok
	case err := <-errChan:
		t.Fatalf("Client connection failed: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Client connection timed out waiting for headers")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	if n == 0 {
		time.Sleep(10 * time.Millisecond)
		n, _ = resp.Body.Read(buf)
	}
	output := string(buf[:n])

	assert.Contains(t, output, "event: error-event")
	// Fallback uses fmt.Fprintf(..., "data: %v\n\n", event.Data)
	assert.Contains(t, output, "data: ")
}
