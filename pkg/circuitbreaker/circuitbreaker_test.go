package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	// Configure aggressive settings for testing
	Configure(true, 3, 1*time.Second, 1*time.Second)

	name := "test-cb"
	errDummy := errors.New("dummy error")

	// 1. Fail 3 times to trip
	for i := 0; i < 3; i++ {
		err := Execute(name, func() error {
			return errDummy
		})
		assert.Equal(t, errDummy, err)
	}

	// 2. Next call should be ErrOpenState (from gobreaker)
	// gobreaker returns "circuit breaker is open" error
	err := Execute(name, func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")

	// 3. Wait for timeout (1s) to allow transition to Half-Open
	time.Sleep(1100 * time.Millisecond)

	// 4. Half-Open -> Success -> Closed
	err = Execute(name, func() error {
		return nil
	})
	assert.NoError(t, err)

	// 5. Subsequent calls should work
	err = Execute(name, func() error {
		return nil
	})
	assert.NoError(t, err)
}
