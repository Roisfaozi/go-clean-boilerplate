package ws

import (
	"context"
	"net/http"
	"time"

	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WebSocketController struct {
	log      *logrus.Logger
	manager  Manager
	upgrader *websocket.Upgrader
	userRepo userRepo.UserRepository
	enforcer *casbin.Enforcer
}

func NewWebSocketController(log *logrus.Logger, manager Manager, allowedOrigins []string, userRepo userRepo.UserRepository, enforcer *casbin.Enforcer) *WebSocketController {
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
		userRepo: userRepo,
		enforcer: enforcer,
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

	userIDVal, exists := ctx.Get("user_id")
	userID := ""
	if exists && userIDVal != nil {
		userID = userIDVal.(string)
	}

	orgIDVal, exists := ctx.Get("organization_id")
	orgID := "global"
	if exists && orgIDVal != nil && orgIDVal != "" {
		orgID = orgIDVal.(string)
	}

	// Fetch User Details for Presence
	var userData *PresenceUser
	if userID != "" && c.userRepo != nil {
		user, err := c.userRepo.FindByID(context.Background(), userID)
		if err == nil && user != nil {
			role := "member"
			if c.enforcer != nil {
				roles, _ := c.enforcer.GetRolesForUser(userID, orgID)
				if len(roles) > 0 {
					role = roles[0]
				}
			}
			userData = &PresenceUser{
				UserID:    userID,
				Name:      user.Name,
				AvatarURL: user.AvatarURL,
				Role:      role,
				Status:    "online",
			}
		}
	}

	client := NewWebsocketClient(conn, c.manager, c.log, config, userID, orgID, userData)

	c.manager.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()

	c.log.Infof("New WebSocket connection established for user %s in org %s", userID, orgID)
}
