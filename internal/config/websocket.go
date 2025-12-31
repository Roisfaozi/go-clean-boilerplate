package config

import (
	"log"
	"net/http"
	"time"

	wsPkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gorilla/websocket"
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
		DistributedEnabled: false,
		RedisPrefix:        "ws_broadcast:",
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

func (c *WebSocketConfig) GetUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			log.Println("WARNING: WebSocket CheckOrigin is permitting all origins. This is unsafe for production.")
			return true
		},
	}
}
