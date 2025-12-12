package sse_test

import (
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	manager := sse.NewManager()
	assert.NotNil(t, manager)
	manager.Stop()
}

func TestRegisterAndUnregisterClient(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event)
	client := &sse.Client{Channel: clientChan}

	manager.RegisterClient(client)
	
	// Unregister triggers channel close
	manager.UnregisterClient(client)
}

func TestBroadcast(t *testing.T) {
	manager := sse.NewManager()
	defer manager.Stop()

	clientChan := make(chan sse.Event, 1)
	client := &sse.Client{Channel: clientChan}
	manager.RegisterClient(client)

	// Allow time for registration to be processed by run loop
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