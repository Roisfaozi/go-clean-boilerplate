package telemetry

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisMetricsHook_DirectHooks(t *testing.T) {
	hook := NewRedisMetricsHook()
	ctx := context.Background()

	t.Run("DialHook delegates directly", func(t *testing.T) {
		called := false
		dial := hook.DialHook(func(ctx context.Context, network, addr string) (net.Conn, error) {
			called = true
			return nil, nil
		})
		conn, err := dial(ctx, "tcp", "127.0.0.1:6379")
		assert.NoError(t, err)
		assert.Nil(t, conn)
		assert.True(t, called)
	})

	t.Run("ProcessHook records ok, miss, error, and unknown commands", func(t *testing.T) {
		beforeSetOK := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("set", "ok"))
		beforeGetMiss := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "miss"))
		beforePingErr := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("ping", "error"))
		beforeOtherOK := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("other", "ok"))

		okProc := hook.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
			return nil
		})
		require.NoError(t, okProc(ctx, redis.NewStatusCmd(ctx, "set", "k", "v")))
		require.NoError(t, okProc(ctx, redis.NewCmd(ctx, "arbitrary_cmd", "arg1")))
		require.NoError(t, okProc(ctx, nil))

		missProc := hook.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
			return redis.Nil
		})
		err := missProc(ctx, redis.NewStringCmd(ctx, "get", "missing"))
		assert.ErrorIs(t, err, redis.Nil)

		boom := errors.New("connection refused")
		errProc := hook.ProcessHook(func(ctx context.Context, cmd redis.Cmder) error {
			return boom
		})
		err = errProc(ctx, redis.NewStatusCmd(ctx, "ping"))
		assert.ErrorIs(t, err, boom)

		assert.Equal(t, beforeSetOK+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("set", "ok")))
		assert.Equal(t, beforeOtherOK+2, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("other", "ok")))
		assert.Equal(t, beforeGetMiss+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "miss")))
		assert.Equal(t, beforePingErr+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("ping", "error")))
	})

	t.Run("ProcessPipelineHook records ok, miss, and error", func(t *testing.T) {
		beforePipeOK := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "ok"))
		beforePipeMiss := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "miss"))
		beforePipeErr := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "error"))

		cmds := []redis.Cmder{redis.NewStringCmd(ctx, "get", "a")}

		okPipe := hook.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
			return nil
		})
		require.NoError(t, okPipe(ctx, cmds))

		missPipe := hook.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
			return redis.Nil
		})
		assert.ErrorIs(t, missPipe(ctx, cmds), redis.Nil)

		boom := errors.New("broken pipe")
		errPipe := hook.ProcessPipelineHook(func(ctx context.Context, cmds []redis.Cmder) error {
			return boom
		})
		assert.ErrorIs(t, errPipe(ctx, cmds), boom)

		assert.Equal(t, beforePipeOK+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "ok")))
		assert.Equal(t, beforePipeMiss+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "miss")))
		assert.Equal(t, beforePipeErr+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "error")))
	})
}

func TestRedisMetricsHook_MiniredisIntegration(t *testing.T) {
	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr:        mr.Addr(),
		DialTimeout: time.Second,
		ReadTimeout: time.Second,
	})
	defer func() { _ = client.Close() }()
	client.AddHook(NewRedisMetricsHook())

	ctx := context.Background()

	beforeSetOK := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("set", "ok"))
	beforeGetOK := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "ok"))
	beforeGetMiss := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "miss"))
	beforePipeOK := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "ok"))
	beforeGetErr := testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "error"))

	require.NoError(t, client.Set(ctx, "k1", "v1", 0).Err())
	val, err := client.Get(ctx, "k1").Result()
	require.NoError(t, err)
	assert.Equal(t, "v1", val)

	err = client.Get(ctx, "missing-key").Err()
	assert.ErrorIs(t, err, redis.Nil)

	pipe := client.Pipeline()
	pipe.Incr(ctx, "counter")
	pipe.Expire(ctx, "counter", time.Minute)
	_, err = pipe.Exec(ctx)
	require.NoError(t, err)

	mr.Close()
	err = client.Get(ctx, "k1").Err()
	require.Error(t, err)
	assert.NotErrorIs(t, err, redis.Nil)

	assert.Equal(t, beforeSetOK+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("set", "ok")))
	assert.Equal(t, beforeGetOK+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "ok")))
	assert.Equal(t, beforeGetMiss+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "miss")))
	assert.Equal(t, beforePipeOK+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("pipeline", "ok")))
	assert.Equal(t, beforeGetErr+1, testutil.ToFloat64(RedisOperationsTotal.WithLabelValues("get", "error")))
}
