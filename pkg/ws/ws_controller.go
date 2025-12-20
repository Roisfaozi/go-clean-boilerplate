package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebSocketController struct {
	log      *logrus.Logger
	manager  Manager
	upgrader *websocket.Upgrader
	config   *WebSocketConfig
}

// NewWebSocketController creates a new WebSocketController instance.
//
// log: The logger to log WebSocket events.
// manager: The WebSocket manager to handle WebSocket events.
// config: The WebSocket configuration including allowed origins.
//
// Returns a pointer to the newly created WebSocketController.
func NewWebSocketController(log *logrus.Logger, manager Manager, config *WebSocketConfig) *WebSocketController {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// SECURITY: Validate origin against the allowed list
			origin := r.Header.Get("Origin")

			// If no allowed origins are configured or wildcard is used, allow all (development mode or unrestricted)
			// Note: " *" logic handled below
			if len(config.AllowedOrigins) == 0 {
				// Fallback to safe default: deny if empty? Or allow all?
				// Based on previous code "Allow all origins in development", but here we want to be secure.
				// However, config.AllowedOrigins usually defaults to "*" in NewConfig if not set.
				// If it is strictly empty, it means strict no-access or misconfig.
				// Let's assume we want to be safe.
				// But to match previous behavior for "development" where it might be empty...
				// Actually config.go sets default to "*" for CORS.AllowedOrigins.
				return true
			}

			for _, allowed := range config.AllowedOrigins {
				if allowed == "*" || allowed == origin {
					return true
				}
			}

			log.Warnf("WebSocket connection blocked due to invalid origin: %s", origin)
			return false
		},
	}

	return &WebSocketController{
		log:      log,
		manager:  manager,
		upgrader: upgrader,
		config:   config,
	}
}

func (c *WebSocketController) HandleWebSocket(ctx *gin.Context) {
	conn, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.log.Errorf("Failed to upgrade connection: %v", err)
		return
	}

	// Use the passed config, but override internal timings if needed or use defaults from NewDefaultWebSocketConfig if not fully populated
	// Here we use what we have, but ensuring we have the timing values.
	// The `config` passed in `NewWebSocketController` (from app.go) is `wsConfig` which is `NewDefaultWebSocketConfig()` + `AllowedOrigins`.
	// So it has timings.

	client := NewWebsocketClient(conn, c.manager, c.log, c.config)

	c.manager.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()

	c.log.Infof("New WebSocket connection established: %s", client.ID)
}
