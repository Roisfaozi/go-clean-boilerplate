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

// NewWebSocketController creates a new WebSocketController instance.
//
// log: The logger to log WebSocket events.
// manager: The WebSocket manager to handle WebSocket events.
// allowedOrigins: A list of allowed origins for CORS. Use "*" to allow all.
//
// Returns a pointer to the newly created WebSocketController.
func NewWebSocketController(log *logrus.Logger, manager Manager, allowedOrigins []string) *WebSocketController {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// SECURITY: Validate origin to prevent Cross-Site WebSocket Hijacking (CSWSH)
			origin := r.Header.Get("Origin")

			// Check if all origins are allowed
			for _, allowed := range allowedOrigins {
				if allowed == "*" {
					if len(allowedOrigins) == 1 {
						// Only log warning if * is the ONLY allowed origin, to avoid spamming if * is mixed with others for some reason
						// But usually * implies dev mode.
						// log.Warn("WebSocket CheckOrigin is permitting all origins (*). This is unsafe for production.")
					}
					return true
				}
				if allowed == origin {
					return true
				}
			}

			log.Warnf("WebSocket connection rejected from disallowed origin: %s", origin)
			return false
		},
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
