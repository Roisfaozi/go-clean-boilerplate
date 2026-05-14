package ws

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupClientTest() (*Client, *MockManager, *MockPresenceManager) {
	mockManager := new(MockManager)
	mockPresence := new(MockPresenceManager)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	config := &WebSocketConfig{}

	client := &Client{
		ID:      "client-1",
		UserID:  "user-1",
		OrgID:   "org-1",
		Manager: mockManager,
		Send:    make(chan []byte, 10),
		Log:     logger,
		Config:  config,
	}

	return client, mockManager, mockPresence
}

func connectMockClient(url string) (*websocket.Conn, error) {
	wsURL := "ws" + strings.TrimPrefix(url, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	return conn, err
}

func TestClient_Pump_Errors(t *testing.T) {
	// Setup test server to get a real websocket connection
	var clientConn *websocket.Conn
	var serverConn *websocket.Conn
	done := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		serverConn, err = upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		close(done)
	}))
	defer server.Close()

	clientConn, err := connectMockClient(server.URL)
	assert.NoError(t, err)
	defer func() { _ = clientConn.Close() }()

	<-done

	client, mockManager, _ := setupClientTest()
	client.Conn = serverConn
	client.Config = &WebSocketConfig{
		WriteWait:      1 * time.Second,
		PongWait:       1 * time.Second,
		PingPeriod:     10 * time.Millisecond,
		MaxMessageSize: 512,
	}

	mockManager.On("UnregisterClient", client).Return()

	// Test ReadPump on closed connection
	_ = serverConn.Close()
	client.ReadPump() // Should return immediately due to error

	// Re-establish connection for WritePump test
	done2 := make(chan struct{})
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		serverConn, err = upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		close(done2)
	}))
	defer server2.Close()

	clientConn2, err := connectMockClient(server2.URL)
	assert.NoError(t, err)
	defer func() { _ = clientConn2.Close() }()
	<-done2

	client.Conn = serverConn
	client.Send <- []byte("test") // Queue a message
	_ = serverConn.Close()        // Close to cause write error
	client.WritePump()            // Should return on error

	// Wait for ping ticker write error
	// Reset connection
	done3 := make(chan struct{})
	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		serverConn, err = upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		close(done3)
	}))
	defer server3.Close()

	clientConn3, err := connectMockClient(server3.URL)
	assert.NoError(t, err)
	defer func() { _ = clientConn3.Close() }()
	<-done3

	client.Conn = serverConn
	// Empty out the queue
	for len(client.Send) > 0 {
		<-client.Send
	}

	// Close connection to cause ping error
	_ = serverConn.Close()
	client.WritePump()
}

func TestClient_HandleMessage_Subscribe(t *testing.T) {
	client, mockManager, _ := setupClientTest()
	channelName := "test-channel"

	// Expect SubscribeToChannel call
	mockManager.On("SubscribeToChannel", client, channelName).Return()

	msg := ClientMessage{
		Type:    "subscribe",
		Channel: channelName,
	}
	payload, _ := json.Marshal(msg)

	client.handleMessage(payload)

	mockManager.AssertExpectations(t)

	// Check response in Send channel
	select {
	case responseBytes := <-client.Send:
		var response ServerMessage
		err := json.Unmarshal(responseBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, "info", response.Type)
		assert.Equal(t, channelName, response.Channel)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestClient_HandleMessage_Unsubscribe(t *testing.T) {
	client, mockManager, _ := setupClientTest()
	channelName := "test-channel"

	// Expect UnsubscribeFromChannel call
	mockManager.On("UnsubscribeFromChannel", client, channelName).Return()

	msg := ClientMessage{
		Type:    "unsubscribe",
		Channel: channelName,
	}
	payload, _ := json.Marshal(msg)

	client.handleMessage(payload)

	mockManager.AssertExpectations(t)

	select {
	case responseBytes := <-client.Send:
		var response ServerMessage
		err := json.Unmarshal(responseBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, "info", response.Type)
		assert.Equal(t, channelName, response.Channel)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for response")
	}
}

func TestClient_HandleMessage_PresenceHeartbeat(t *testing.T) {
	client, mockManager, mockPresence := setupClientTest()

	// Expect GetPresenceManager call
	mockManager.On("GetPresenceManager").Return(mockPresence)
	// Expect RefreshUserHeartbeat call
	mockPresence.On("RefreshUserHeartbeat", mock.Anything, client.OrgID, client.UserID).Return(nil)

	msg := ClientMessage{
		Type: "presence_heartbeat",
	}
	payload, _ := json.Marshal(msg)

	client.handleMessage(payload)

	mockManager.AssertExpectations(t)
	mockPresence.AssertExpectations(t)
	// Heartbeat does not send response
	assert.Empty(t, client.Send)
}

func TestClient_HandleMessage_Message(t *testing.T) {
	client, mockManager, _ := setupClientTest()
	msg := ClientMessage{
		Type:    "message",
		Channel: "test",
		Data:    []byte(`"hello"`),
	}
	payload, _ := json.Marshal(msg)

	client.handleMessage(payload)

	mockManager.AssertNotCalled(t, "SubscribeToChannel")
}

func TestClient_HandleMessage_UnknownType(t *testing.T) {
	client, mockManager, _ := setupClientTest()

	msg := ClientMessage{
		Type: "unknown_type",
	}
	payload, _ := json.Marshal(msg)

	client.handleMessage(payload)

	mockManager.AssertNotCalled(t, "SubscribeToChannel")

	select {
	case responseBytes := <-client.Send:
		var response ServerMessage
		err := json.Unmarshal(responseBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Type)
		assert.Contains(t, response.Data, "Unknown message type")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for error response")
	}
}

func TestClient_HandleMessage_InvalidJSON(t *testing.T) {
	client, mockManager, _ := setupClientTest()

	client.handleMessage([]byte("invalid-json"))

	mockManager.AssertNotCalled(t, "SubscribeToChannel")

	select {
	case responseBytes := <-client.Send:
		var response ServerMessage
		err := json.Unmarshal(responseBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Type)
		assert.Contains(t, response.Data, "Invalid message format")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for error response")
	}
}

func TestClient_SendJSON_BufferFull(t *testing.T) {
	client, _, _ := setupClientTest()
	// Fill the buffer
	for i := 0; i < 10; i++ {
		client.Send <- []byte("msg")
	}

	// Try to send one more
	client.sendJSON(map[string]string{"test": "full"})

	// Should not block and should log warning (cannot assert log easily without hook, but ensures no deadlock)
	// Channel should be full
	assert.Equal(t, 10, len(client.Send))
}

func TestClient_SendJSON_MarshalError(t *testing.T) {
	client, _, _ := setupClientTest()

	// Send channel (func) which cannot be marshaled
	client.sendJSON(func() {})

	// Should not panic, should log error
	assert.Equal(t, 0, len(client.Send))
}
