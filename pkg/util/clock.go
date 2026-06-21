package util

import "time"

// Clock is an interface for time operations, allowing for easier testing.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// RealClock implements Clock using the standard time package.
type RealClock struct{}

func (RealClock) Now() time.Time                         { return time.Now() }
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// MockClock implements Clock for testing purposes with a fixed time.
type MockClock struct {
	CurrentTime time.Time
}

func (m *MockClock) Now() time.Time                         { return m.CurrentTime }
func (m *MockClock) After(d time.Duration) <-chan time.Time { return make(chan time.Time) }

func (m *MockClock) SetTime(t time.Time) {
	m.CurrentTime = t
}

func (m *MockClock) Add(d time.Duration) {
	m.CurrentTime = m.CurrentTime.Add(d)
}
