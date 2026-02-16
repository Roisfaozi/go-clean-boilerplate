package ws

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockManager is a mock implementation of Manager
type MockManager struct {
	mock.Mock
}

func (m *MockManager) Run() {
	m.Called()
}

func (m *MockManager) RegisterClient(client *Client) {
	m.Called(client)
}

func (m *MockManager) UnregisterClient(client *Client) {
	m.Called(client)
}

func (m *MockManager) BroadcastToChannel(channel string, message []byte) {
	m.Called(channel, message)
}

func (m *MockManager) SubscribeToChannel(client *Client, channel string) {
	m.Called(client, channel)
}

func (m *MockManager) UnsubscribeFromChannel(client *Client, channel string) {
	m.Called(client, channel)
}

func (m *MockManager) GetChannelClients(channel string) int {
	args := m.Called(channel)
	return args.Int(0)
}

func (m *MockManager) PresenceUpdate(orgID string, event string, userData *PresenceUser) {
	m.Called(orgID, event, userData)
}

func (m *MockManager) GetPresenceManager() PresenceManager {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(PresenceManager)
}

// MockPresenceManager is a mock implementation of PresenceManager
type MockPresenceManager struct {
	mock.Mock
}

func (m *MockPresenceManager) SetUserOnline(ctx context.Context, orgID, userID string, userData *PresenceUser) error {
	args := m.Called(ctx, orgID, userID, userData)
	return args.Error(0)
}

func (m *MockPresenceManager) SetUserOffline(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockPresenceManager) GetOnlineUsers(ctx context.Context, orgID string) ([]PresenceUser, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PresenceUser), args.Error(1)
}

func (m *MockPresenceManager) RefreshUserHeartbeat(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockPresenceManager) PruneStaleUsers(ctx context.Context, timeout time.Duration) (map[string][]string, error) {
	args := m.Called(ctx, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]string), args.Error(1)
}

// TestMock just to satisfy coverage if needed, or to ensure mocks are compilable
func TestMocks(t *testing.T) {
	// No-op
}
