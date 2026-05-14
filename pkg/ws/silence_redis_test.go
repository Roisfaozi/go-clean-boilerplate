package ws_test

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type silentLogger struct{}

func (l silentLogger) Printf(ctx context.Context, format string, v ...interface{}) {}

func init() {
	redis.SetLogger(silentLogger{})
}
