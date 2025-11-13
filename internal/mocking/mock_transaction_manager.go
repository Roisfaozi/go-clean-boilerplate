package mocking

import (
	"context"
	"github.com/stretchr/testify/mock"
)

// MockTransactionManager is a mocking implementation of the TransactionManager interface
type MockTransactionManager struct {
	mock.Mock
}

// WithinTransaction mocking the WithinTransaction method
func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	if args.Get(0) == nil {
		return fn(ctx)
	}
	return args.Error(0)
}

// NewMockTransactionManager creates a new instance of MockTransactionManager
func NewMockTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{}
}
