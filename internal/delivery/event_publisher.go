package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/sirupsen/logrus"
)

type eventPublisher struct {
	wsManager  ws.Manager
	sseManager *sse.Manager
	log        *logrus.Logger
}

func NewEventPublisher(wsManager ws.Manager, sseManager *sse.Manager, log *logrus.Logger) repository.NotificationPublisher {
	return &eventPublisher{
		wsManager:  wsManager,
		sseManager: sseManager,
		log:        log,
	}
}

func (p *eventPublisher) PublishUserLoggedIn(ctx context.Context, user model.UserInfo, orgIDs []string) {
	notification := map[string]string{
		"type":    "user_login",
		"user_id": user.ID,
		"message": fmt.Sprintf("User '%s' has just logged in.", user.Name),
		"time":    time.Now().Format(time.RFC3339),
	}
	notificationJSON, _ := json.Marshal(notification)

	if p.wsManager != nil {
		for _, orgID := range orgIDs {
			channel := fmt.Sprintf("org_%s_notifications", orgID)
			p.wsManager.BroadcastToChannel(channel, notificationJSON)
		}
	}

	if p.sseManager != nil {
		p.sseManager.Broadcast("user_login", notification)
	}
}
