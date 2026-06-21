package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockTransactionManager is a mock implementation of WithTransactionManager
type MockTransactionManager struct {
	mock.Mock
}

// WithinTransaction executes the function within a transaction context (mocked)
func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}
