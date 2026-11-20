package worker

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstrumentTaskHandler(t *testing.T) {
	t.Run("records metrics on success", func(t *testing.T) {
		handler := InstrumentTaskHandler("test.task", func(ctx context.Context, task *asynq.Task) error {
			return nil
		})

		err := handler(context.Background(), asynq.NewTask("test.task", nil))
		assert.NoError(t, err)
	})

	t.Run("records metrics on error and propagates error", func(t *testing.T) {
		expectedErr := errors.New("task failed")
		handler := InstrumentTaskHandler("test.task", func(ctx context.Context, task *asynq.Task) error {
			return expectedErr
		})

		err := handler(context.Background(), asynq.NewTask("test.task", nil))
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestRedisTaskProcessor_SetRedisClient(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	tp := NewRedisTaskProcessor(asynq.RedisClientOpt{Addr: "localhost:6379"}, logger, nil, nil, nil, nil, WorkerConfig{})
	proc, ok := tp.(*RedisTaskProcessor)
	require.True(t, ok)

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer func() { _ = rdb.Close() }()

	proc.SetRedisClient(rdb)
	assert.Same(t, rdb, proc.redisClient)
}
