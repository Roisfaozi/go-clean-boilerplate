package config

import (
	"time"
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
	pongWait := 60 * time.Second
	return &WebSocketConfig{
		WriteWait:          10 * time.Second,
		PongWait:           pongWait,
		PingPeriod:         (pongWait * 9) / 10,
		MaxMessageSize:     512 * 1024,
		DistributedEnabled: false, // Disabled by default
		RedisPrefix:        "ws_broadcast:",
	}
}
