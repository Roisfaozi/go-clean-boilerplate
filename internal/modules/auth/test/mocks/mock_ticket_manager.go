package mocks

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/stretchr/testify/mock"
)

type MockTicketManager struct {
	mock.Mock
}

func (m *MockTicketManager) CreateTicket(ctx context.Context, userID, orgID, sessionID, role, username string) (string, error) {
	args := m.Called(ctx, userID, orgID, sessionID, role, username)
	return args.String(0), args.Error(1)
}

func (m *MockTicketManager) ValidateTicket(ctx context.Context, ticket string) (*ws.UserContext, error) {
	args := m.Called(ctx, ticket)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ws.UserContext), args.Error(1)
}
