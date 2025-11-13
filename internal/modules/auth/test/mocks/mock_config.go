package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockConfig is a mock implementation of the usecase.Config interface
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetAccessTokenSecret() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfig) GetRefreshTokenSecret() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConfig) GetAccessTokenDuration() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetRefreshTokenDuration() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}
