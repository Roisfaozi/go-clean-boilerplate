package ws

import (
	"encoding/json"
	"io"
	"testing"
	"time"

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
