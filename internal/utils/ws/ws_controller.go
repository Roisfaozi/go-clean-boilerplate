package ws

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WebSocketController handles WebSocket connections
type WebSocketController struct {
	log      *logrus.Logger
	manager  Manager
	upgrader *websocket.Upgrader
}

// NewWebSocketController creates a new WebSocket controller
func NewWebSocketController(log *logrus.Logger, manager Manager) *WebSocketController {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins in development
			// In production, you should validate the origin
			return true
		},
	}

	return &WebSocketController{
		log:      log,
		manager:  manager,
		upgrader: upgrader,
	}
}

// HandleWebSocket handles WebSocket connection requests
func (c *WebSocketController) HandleWebSocket(ctx *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.log.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Get WebSocket Config from Manager
	config := &WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     54 * time.Second,
		MaxMessageSize: 512 * 1024,
	}

	// Create new client
	client := NewClient(conn, c.manager, c.log, config)

	// Register client with Manager
	c.manager.RegisterClient(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()

	c.log.Infof("New WebSocket connection established: %s", client.ID)
}
