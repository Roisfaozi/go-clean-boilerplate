package telemetry

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisMetricsHook struct{}

// NewRedisMetricsHook returns a go-redis Hook that records bounded command latency and status metrics.
func NewRedisMetricsHook() redis.Hook {
	return &redisMetricsHook{}
}

func (h *redisMetricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *redisMetricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		cmdName := ""
		if cmd != nil {
			cmdName = cmd.Name()
		}
		RecordRedisOperation(cmdName, classifyRedisError(err), time.Since(start))
		return err
	}
}

func (h *redisMetricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		RecordRedisOperation("pipeline", classifyRedisError(err), time.Since(start))
		return err
	}
}

func classifyRedisError(err error) string {
	if err == nil {
		return "ok"
	}
	if errors.Is(err, redis.Nil) {
		return "miss"
	}
	return "error"
}
