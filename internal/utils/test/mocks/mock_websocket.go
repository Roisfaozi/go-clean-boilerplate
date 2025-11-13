package mocks

import (
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketManager is a mock implementation of the ws.Manager interface
type MockWebSocketManager struct {
	mock.Mock
}

func (m *MockWebSocketManager) Run() {
	m.Called()
}

func (m *MockWebSocketManager) RegisterClient(client *ws.Client) {
	m.Called(client)
}

func (m *MockWebSocketManager) UnregisterClient(client *ws.Client) {
	m.Called(client)
}

func (m *MockWebSocketManager) BroadcastToChannel(channel string, message []byte) {
	m.Called(channel, message)
}

func (m *MockWebSocketManager) SubscribeToChannel(client *ws.Client, channel string) {
	m.Called(client, channel)
}

func (m *MockWebSocketManager) UnsubscribeFromChannel(client *ws.Client, channel string) {
	m.Called(client, channel)
}

func (m *MockWebSocketManager) GetChannelClients(channel string) int {
	args := m.Called(channel)
	return args.Int(0)
}

// NoOpWriter is a simple io.Writer that does nothing.
// It's useful for suppressing log output during tests.
type NoOpWriter struct{}

// Write implements the io.Writer interface.
func (w *NoOpWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}
