package config

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebSocketConfig struct {
	WriteWait      time.Duration `mapstructure:"write_wait"`
	PongWait       time.Duration `mapstructure:"pong_wait"`
	PingPeriod     time.Duration `mapstructure:"ping_period"`
	MaxMessageSize int64         `mapstructure:"max_message_size"`
	AllowedOrigins []string      `mapstructure:"allowed_origins"`
}

func NewDefaultWebSocketConfig() *WebSocketConfig {
	pongWait := 60 * time.Second
	return &WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       pongWait,
		PingPeriod:     (pongWait * 9) / 10,
		MaxMessageSize: 512 * 1024,
	}
}

func (c *WebSocketConfig) GetUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// SECURITY: In a production environment, you MUST implement a proper origin check
			// to prevent Cross-Site WebSocket Hijacking (CSWSH).
			// Example:
			// origin := r.Header.Get("Origin")
			// return origin == "https://your-allowed-domain.com"

			// For this boilerplate/demo, we log a warning but allow it.
			// TODO: Configure allowed origins via config.
			log.Println("WARNING: WebSocket CheckOrigin is permitting all origins. This is unsafe for production.")
			return true
		},
	}
}
