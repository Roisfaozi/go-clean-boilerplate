package config

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	WriteWait      time.Duration `mapstructure:"write_wait"`
	PongWait       time.Duration `mapstructure:"pong_wait"`
	PingPeriod     time.Duration `mapstructure:"ping_period"`
	MaxMessageSize int64         `mapstructure:"max_message_size"`
}

// NewDefaultWebSocketConfig creates a WebSocket configuration with default values
func NewDefaultWebSocketConfig() *WebSocketConfig {
	// The PingPeriod must be less than the PongWait.
	pongWait := 60 * time.Second
	return &WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       pongWait,
		PingPeriod:     (pongWait * 9) / 10, // Recommended to be less than PongWait
		MaxMessageSize: 512 * 1024,          // 512KB
	}
}

// GetUpgrader returns a configured WebSocket upgrader
func (c *WebSocketConfig) GetUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// In a production environment, you should implement a proper origin check.
			// For example:
			// origin := r.Header.Get("Origin")
			// return origin == "https://your-allowed-domain.com"
			return true // Allow all origins for development purposes
		},
	}
}
