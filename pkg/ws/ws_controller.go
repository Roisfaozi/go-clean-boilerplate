package ws

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebSocketController struct {
	log      *logrus.Logger
	manager  Manager
	upgrader *websocket.Upgrader
}

func NewWebSocketController(log *logrus.Logger, manager Manager, allowedOrigins []string) *WebSocketController {
	checkOrigin := func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				return true
			}
		}
		return false
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}

	if len(allowedOrigins) == 0 {
		upgrader.CheckOrigin = nil
	}

	return &WebSocketController{
		log:      log,
		manager:  manager,
		upgrader: upgrader,
	}
}

func (c *WebSocketController) HandleWebSocket(ctx *gin.Context) {
	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.log.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	config := &WebSocketConfig{
		WriteWait:      10 * time.Second,
		PongWait:       60 * time.Second,
		PingPeriod:     54 * time.Second,
		MaxMessageSize: 512 * 1024,
	}

	client := NewWebsocketClient(conn, c.manager, c.log, config)

	c.manager.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()

	c.log.Infof("New WebSocket connection established: %s", client.ID)
}
