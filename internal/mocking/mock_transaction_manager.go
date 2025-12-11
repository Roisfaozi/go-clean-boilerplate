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

	// If an error is configured to be returned by the mock, return it immediately.
	if err := args.Error(0); err != nil {
		return err
	}

	// Otherwise, execute the function that was passed in.
	return fn(ctx)
}

// NewMockTransactionManager creates a new instance of MockTransactionManager
func NewMockTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{}
}
