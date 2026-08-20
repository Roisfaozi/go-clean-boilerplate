package config

import (
	"time"

	wsPkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
)

const (
	defaultWebSocketWriteWait            = 10 * time.Second
	defaultWebSocketPongWait             = 60 * time.Second
	defaultWebSocketPingPeriod           = (defaultWebSocketPongWait * 9) / 10
	defaultWebSocketMaxMessageSize int64 = 512 * 1024
)

type WebSocketConfig struct {
	WriteWait          time.Duration `mapstructure:"write_wait"`
	PongWait           time.Duration `mapstructure:"pong_wait"`
	PingPeriod         time.Duration `mapstructure:"ping_period"`
	MaxMessageSize     int64         `mapstructure:"max_message_size"`
	DistributedEnabled bool          `mapstructure:"distributed_enabled"`
	RedisPrefix        string        `mapstructure:"redis_prefix"`
}

func NewDefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		WriteWait:          defaultWebSocketWriteWait,
		PongWait:           defaultWebSocketPongWait,
		PingPeriod:         defaultWebSocketPingPeriod,
		MaxMessageSize:     defaultWebSocketMaxMessageSize,
		DistributedEnabled: false,
		RedisPrefix:        defaultWebSocketRedisPrefix,
	}
}

// ToPkgConfig maps internal config to package-level config safely.
func (c *WebSocketConfig) ToPkgConfig() *wsPkg.WebSocketConfig {
	return &wsPkg.WebSocketConfig{
		WriteWait:          c.WriteWait,
		PongWait:           c.PongWait,
		PingPeriod:         c.PingPeriod,
		MaxMessageSize:     c.MaxMessageSize,
		DistributedEnabled: c.DistributedEnabled,
		RedisPrefix:        c.RedisPrefix,
	}
}
