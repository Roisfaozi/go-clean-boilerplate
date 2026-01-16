package circuitbreaker

import (
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

type Settings struct {
	Enabled     bool
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
}

var (
	breakers = make(map[string]*gobreaker.CircuitBreaker)
	mu       sync.RWMutex
	defaultSettings Settings
)

// Configure sets global settings for new breakers
func Configure(enabled bool, maxRequests uint32, interval, timeout time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	defaultSettings = Settings{
		Enabled:     enabled,
		MaxRequests: maxRequests,
		Interval:    interval,
		Timeout:     timeout,
	}
}

// GetOrBuild returns an existing breaker or creates a new one with global settings
func GetOrBuild(name string) *gobreaker.CircuitBreaker {
	mu.RLock()
	cb, exists := breakers[name]
	mu.RUnlock()

	if exists {
		return cb
	}

	mu.Lock()
	defer mu.Unlock()

	// Double check
	if cb, exists = breakers[name]; exists {
		return cb
	}


st := gobreaker.Settings{
		Name:        name,
		MaxRequests: defaultSettings.MaxRequests,
		Interval:    defaultSettings.Interval,
		Timeout:     defaultSettings.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker '%s' changed state from %s to %s\n", name, from, to)
		},
	}

	cb = gobreaker.NewCircuitBreaker(st)
	breakers[name] = cb
	return cb
}

// Execute runs the given function within the named circuit breaker
func Execute(name string, fn func() error) error {
	mu.RLock()
	enabled := defaultSettings.Enabled
	mu.RUnlock()

	if !enabled {
		return fn()
	}

	cb := GetOrBuild(name)
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, fn()
	})
	return err
}