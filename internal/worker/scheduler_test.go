package worker_test

import (
	"testing"
	"time"
	"os"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSchedulerLifecycle(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}

	scheduler := worker.NewScheduler(redisOpt, logger)

	// Test Register
	scheduler.RegisterScheduledTasks()

	// Test Start and Shutdown
	go func() {
		_ = scheduler.Start()
	}()
	time.Sleep(100 * time.Millisecond)
	scheduler.Shutdown()
	assert.NotNil(t, scheduler)
}
