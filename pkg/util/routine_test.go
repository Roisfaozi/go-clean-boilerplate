package util

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSafeGo_RecoversPanic(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := logrus.New()
	logger.SetOutput(buf)

	var wg sync.WaitGroup
	wg.Add(1)

	SafeGo(logger, func() {
		go func() {
			time.Sleep(10 * time.Millisecond)
			wg.Done()
		}()
		panic("simulated panic")
	})

	wg.Wait()
	assert.Contains(t, buf.String(), "panic recovered in goroutine: simulated panic")
}
