package config

import (
	nethttp "net/http"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

// NewWebSocketConfig creates a new WebSocket configuration from viper
func NewWebSocketConfig(config *viper.Viper) *WebSocketConfig {
	return &WebSocketConfig{
		WriteWait:      config.GetDuration("websocket.write_wait") * time.Second,
		PongWait:       config.GetDuration("websocket.pong_wait") * time.Second,
		PingPeriod:     config.GetDuration("websocket.ping_period") * time.Second,
		MaxMessageSize: config.GetInt64("websocket.max_message_size"),
	}
}

// NewDefaultWebSocketConfig creates a WebSocket configuration with default values
func NewDefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     54 * time.Second,
		MaxMessageSize: 512 * 1024, // 512KB
	}
}

// GetUpgrader returns a configured WebSocket upgrader
func (c *WebSocketConfig) GetUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *nethttp.Request) bool {
			return true // Allow all origins in development
		},
	}
}

// NewWebSocketManager creates a new WebSocket manager with config
func NewWebSocketManager(config *viper.Viper, log *logrus.Logger) ws.Manager {
	var wsConfig *ws.WebSocketConfig

	if config != nil {
		wsConfig = &ws.WebSocketConfig{
			WriteWait:      config.GetDuration("websocket.write_wait") * time.Second,
			PongWait:       config.GetDuration("websocket.pong_wait") * time.Second,
			PingPeriod:     config.GetDuration("websocket.ping_period") * time.Second,
			MaxMessageSize: config.GetInt64("websocket.max_message_size"),
		}
	} else {
		// Default configuration
		wsConfig = &ws.WebSocketConfig{
			WriteWait:      10 * time.Second,
			PongWait:       60 * time.Second,
			PingPeriod:     54 * time.Second,
			MaxMessageSize: 512 * 1024, // 512KB
		}
	}

	return ws.NewWebSocketManager(wsConfig, log)
}
