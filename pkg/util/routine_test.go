package util

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestSafeGo_RecoversPanic(t *testing.T) {
	buf := &safeBuffer{}
	logger := logrus.New()
	logger.SetOutput(buf)

	done := make(chan struct{})

	SafeGo(logger, func() {
		defer close(done)
		panic("simulated panic")
	})

	<-done
	// Allow small window for logger write inside recover defer
	time.Sleep(20 * time.Millisecond)
	assert.Contains(t, buf.String(), "panic recovered in goroutine: simulated panic")
}
