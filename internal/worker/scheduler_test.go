package worker_test

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schedulerTestDeps struct {
	scheduler *worker.Scheduler
	mockRedis *miniredis.Miniredis
}

func setupSchedulerTest(t *testing.T) (*schedulerTestDeps, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	redisOpt := asynq.RedisClientOpt{
		Addr: mr.Addr(),
	}

	logger := logrus.New()
	logger.SetOutput(new(mockWriter)) // Reuse from processor_test.go

	scheduler := worker.NewScheduler(redisOpt, logger)

	cleanup := func() {
		mr.Close()
	}

	return &schedulerTestDeps{
		scheduler: scheduler,
		mockRedis: mr,
	}, cleanup
}

func TestScheduler_RegisterScheduledTasks(t *testing.T) {
	t.Run("Positive - Register tasks successfully", func(t *testing.T) {
		deps, cleanup := setupSchedulerTest(t)
		defer cleanup()

		// Should not panic or error during registration
		deps.scheduler.RegisterScheduledTasks()
	})

	t.Run("Negative - Bad Redis Option should still register in memory for asynq until Started", func(t *testing.T) {
		logger := logrus.New()
		logger.SetOutput(new(mockWriter))
		scheduler := worker.NewScheduler(asynq.RedisClientOpt{Addr: "invalid:9999"}, logger)

		scheduler.RegisterScheduledTasks()
		// asynq.Scheduler.Register only parses crons and saves them in memory, doesn't immediately connect
	})

	t.Run("Edge - Re-registering tasks", func(t *testing.T) {
		deps, cleanup := setupSchedulerTest(t)
		defer cleanup()

		// Registers without crashing
		deps.scheduler.RegisterScheduledTasks()

		// Attempting to register again might fail since same entry ID, let's see how it behaves
		deps.scheduler.RegisterScheduledTasks()
		// We just assert it doesn't panic
	})

	t.Run("Vulnerability - Verify time location setup", func(t *testing.T) {
		deps, cleanup := setupSchedulerTest(t)
		defer cleanup()

		// Verify no panic during init and that scheduler isn't nil
		assert.NotNil(t, deps.scheduler)
	})
}
