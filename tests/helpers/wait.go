package helpers

import (
	"fmt"
	"time"
)

func WaitForCondition(condition func() bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("condition not met within timeout of %v", timeout)
}

func WaitForConditionWithRetry(condition func() (bool, error), timeout time.Duration, retryInterval time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ok, err := condition()
		if err != nil {
			return fmt.Errorf("condition check failed: %w", err)
		}
		if ok {
			return nil
		}
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("condition not met within timeout of %v", timeout)
}

func RetryOperation(operation func() error, maxRetries int, retryDelay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}
