package mocks

import (
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockWebSocketManager) PresenceUpdate(orgID string, event string, userData *ws.PresenceUser) {
	m.Called(orgID, event, userData)
}

func (m *MockWebSocketManager) GetPresenceManager() ws.PresenceManager {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(ws.PresenceManager)
}
