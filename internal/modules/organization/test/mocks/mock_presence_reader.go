package mocks

import (
	"context"

	wsPkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/stretchr/testify/mock"
)

// MockPresenceReader is a mock implementation of PresenceReader
type MockPresenceReader struct {
	mock.Mock
}

func (m *MockPresenceReader) GetOnlineUsers(ctx context.Context, orgID string) ([]wsPkg.PresenceUser, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]wsPkg.PresenceUser), args.Error(1)
}
