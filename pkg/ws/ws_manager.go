package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Manager interface {
	Run()
	RegisterClient(client *Client)
	UnregisterClient(client *Client)
	BroadcastToChannel(channel string, message []byte)
	SubscribeToChannel(client *Client, channel string)
	UnsubscribeFromChannel(client *Client, channel string)
	GetChannelClients(channel string) int
	PresenceUpdate(orgID string, event string, userData *PresenceUser)
	GetPresenceManager() PresenceManager
}

type WebSocketManager struct {
	clients map[*Client]bool

	channels map[string]map[*Client]bool

	broadcast chan *BroadcastMessage

	register chan *Client

	unregister chan *Client

	subscribe chan *SubscriptionRequest

	unsubscribe chan *SubscriptionRequest

	mu sync.RWMutex

	log *logrus.Logger

	config *WebSocketConfig

	stopChan chan struct{}

	redisClient *redis.Client

	presence PresenceManager
}

type BroadcastMessage struct {
	Channel    string
	Message    []byte
	FromRemote bool
}

type SubscriptionRequest struct {
	Client  *Client
	Channel string
}

type WebSocketConfig struct {
	WriteWait          time.Duration
	PongWait           time.Duration
	PingPeriod         time.Duration
	MaxMessageSize     int64
	DistributedEnabled bool
	RedisPrefix        string
}

func NewWebSocketManager(config *WebSocketConfig, log *logrus.Logger, redisClient *redis.Client, presence PresenceManager) *WebSocketManager {
	return &WebSocketManager{
		clients:     make(map[*Client]bool),
		channels:    make(map[string]map[*Client]bool),
		broadcast:   make(chan *BroadcastMessage, 256),
		register:    make(chan *Client, 256),
		unregister:  make(chan *Client, 256),
		subscribe:   make(chan *SubscriptionRequest, 256),
		unsubscribe: make(chan *SubscriptionRequest, 256),
		log:         log,
		config:      config,
		stopChan:    make(chan struct{}),
		redisClient: redisClient,
		presence:    presence,
	}
}

func (m *WebSocketManager) Run() {
	m.log.Info("WebSocket Manager started")

	if m.config.DistributedEnabled && m.redisClient != nil {
		go m.listenToRedis()
	}

	for {
		select {
		case <-m.stopChan:
			m.log.Info("WebSocket Manager stopped")
			return

		case client := <-m.register:
			m.handleRegister(client)

		case client := <-m.unregister:
			m.handleUnregister(client)

		case message := <-m.broadcast:
			m.handleBroadcast(message)

		case req := <-m.subscribe:
			m.handleSubscribe(req)

		case req := <-m.unsubscribe:
			m.handleUnsubscribe(req)
		}
	}
}

func (m *WebSocketManager) listenToRedis() {
	ctx := context.Background()
	prefix := m.config.RedisPrefix
	if prefix == "" {
		prefix = "ws_broadcast:"
	}

	pubsub := m.redisClient.PSubscribe(ctx, prefix+"*")
	defer func() {
		_ = pubsub.Close()
	}()

	ch := pubsub.Channel()
	m.log.Infof("Listening to Redis Pub/Sub pattern: %s*", prefix)

	for {
		select {
		case <-m.stopChan:
			return
		case msg := <-ch:

			localChannel := msg.Channel[len(prefix):]

			m.broadcast <- &BroadcastMessage{
				Channel:    localChannel,
				Message:    []byte(msg.Payload),
				FromRemote: true,
			}
		}
	}
}

func (m *WebSocketManager) handleRegister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client] = true
	telemetry.ActiveWSConnections.Inc()
	m.log.Infof("Client registered: %s, total clients: %d", client.ID, len(m.clients))

	// Track Presence if user info is available
	if client.UserID != "" && client.OrgID != "" {
		userData := client.UserData
		if userData == nil {
			userData = &PresenceUser{
				UserID: client.UserID,
				Status: "online",
			}
		}
		
		if err := m.presence.SetUserOnline(context.Background(), client.OrgID, client.UserID, userData); err != nil {
			m.log.WithError(err).Error("Failed to set user online in presence manager")
		} else {
			// Broadcast Join Event
			m.PresenceUpdate(client.OrgID, "join", userData)
		}
	}
}

func (m *WebSocketManager) handleUnregister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[client]; ok {
		// Update Presence
		if client.UserID != "" && client.OrgID != "" {
			// Only set offline if no other connections for this user?
			// For simplicity in MVP, we just set offline.
			// Better: Reference counting or checking other clients.
			if err := m.presence.SetUserOffline(context.Background(), client.OrgID, client.UserID); err != nil {
				m.log.WithError(err).Error("Failed to set user offline in presence manager")
			} else {
				m.PresenceUpdate(client.OrgID, "leave", &PresenceUser{UserID: client.UserID})
			}
		}

		for channel, clients := range m.channels {
			if _, exists := clients[client]; exists {
				delete(clients, client)
				m.log.Infof("Client %s removed from channel: %s", client.ID, channel)

				if len(clients) == 0 {
					delete(m.channels, channel)
					m.log.Infof("Channel removed (empty): %s", channel)
				}
			}
		}

		delete(m.clients, client)
		close(client.Send)

		telemetry.ActiveWSConnections.Dec()
		m.log.Infof("Client unregistered: %s, total clients: %d", client.ID, len(m.clients))
	}
}

func (m *WebSocketManager) handleBroadcast(msg *BroadcastMessage) {

	if !msg.FromRemote && m.config.DistributedEnabled && m.redisClient != nil {
		ctx := context.Background()
		prefix := m.config.RedisPrefix
		if prefix == "" {
			prefix = "ws_broadcast:"
		}

		err := m.redisClient.Publish(ctx, prefix+msg.Channel, msg.Message).Err()
		if err != nil {
			m.log.Errorf("Failed to publish to Redis for channel %s: %v", msg.Channel, err)
		}
		// In distributed mode, we rely on Redis Pub/Sub to echo the message back to us (and other nodes).
		// So we return here to prevent sending the message twice (once locally, once via Redis echo).
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Wrap message in ServerMessage envelope
	envelope := map[string]interface{}{
		"type":    "message",
		"channel": msg.Channel,
		// We try to unmarshal if it looks like JSON, otherwise send as string
		// But to be safe and generic, we can just send it as raw json.RawMessage if we could,
		// but here msg.Message is []byte.
		// Let's assume msg.Message is a JSON string.
		"data": json.RawMessage(msg.Message),
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		m.log.Errorf("Failed to marshal broadcast envelope: %v", err)
		return
	}

	if clients, ok := m.channels[msg.Channel]; ok {
		count := 0
		for client := range clients {
			select {
			case client.Send <- payload:
				count++
			default:
				m.log.Warnf("Failed to Send message to client %s (buffer full)", client.ID)
			}
		}
		m.log.Debugf("Local broadcast to channel %s: %d/%d clients", msg.Channel, count, len(clients))
	}
}

func (m *WebSocketManager) handleSubscribe(req *SubscriptionRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if client is still valid/registered
	if _, registered := m.clients[req.Client]; !registered {
		m.log.Warnf("Client %s tried to subscribe but is not registered", req.Client.ID)
		return
	}

	if _, ok := m.channels[req.Channel]; !ok {
		m.channels[req.Channel] = make(map[*Client]bool)
		m.log.Infof("Channel created: %s", req.Channel)
	}

	m.channels[req.Channel][req.Client] = true
	m.log.Infof("Client %s subscribed to channel: %s, total subscribers: %d",
		req.Client.ID, req.Channel, len(m.channels[req.Channel]))
}

func (m *WebSocketManager) handleUnsubscribe(req *SubscriptionRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, ok := m.channels[req.Channel]; ok {
		if _, exists := clients[req.Client]; exists {
			delete(clients, req.Client)
			m.log.Infof("Client %s unsubscribed from channel: %s, remaining subscribers: %d",
				req.Client.ID, req.Channel, len(clients))

			if len(clients) == 0 {
				delete(m.channels, req.Channel)
				m.log.Infof("Channel removed (empty): %s", req.Channel)
			}
		}
	}
}

func (m *WebSocketManager) RegisterClient(client *Client) {
	select {
	case m.register <- client:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("RegisterClient timed out")
	case <-m.stopChan:
		m.log.Warn("RegisterClient called on stopped manager")
	}
}

func (m *WebSocketManager) UnregisterClient(client *Client) {
	select {
	case m.unregister <- client:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("UnregisterClient timed out")
	case <-m.stopChan:
		m.log.Warn("UnregisterClient called on stopped manager")
	}
}

func (m *WebSocketManager) BroadcastToChannel(channel string, message []byte) {
	select {
	case m.broadcast <- &BroadcastMessage{Channel: channel, Message: message, FromRemote: false}:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("BroadcastToChannel timed out")
	case <-m.stopChan:
		m.log.Warn("BroadcastToChannel called on stopped manager")
	}
}

func (m *WebSocketManager) SubscribeToChannel(client *Client, channel string) {
	select {
	case m.subscribe <- &SubscriptionRequest{Client: client, Channel: channel}:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("SubscribeToChannel timed out")
	case <-m.stopChan:
		m.log.Warn("SubscribeToChannel called on stopped manager")
	}
}

func (m *WebSocketManager) UnsubscribeFromChannel(client *Client, channel string) {
	select {
	case m.unsubscribe <- &SubscriptionRequest{Client: client, Channel: channel}:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("UnsubscribeFromChannel timed out")
	case <-m.stopChan:
		m.log.Warn("UnsubscribeFromChannel called on stopped manager")
	}
}

func (m *WebSocketManager) GetChannelClients(channel string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, ok := m.channels[channel]; ok {
		return len(clients)
	}
	return 0
}

func (m *WebSocketManager) PresenceUpdate(orgID string, event string, userData *PresenceUser) {
	channel := "presence:org:" + orgID
	payload, _ := json.Marshal(map[string]interface{}{
		"event": event,
		"user":  userData,
	})

	m.BroadcastToChannel(channel, payload)
}

func (m *WebSocketManager) GetPresenceManager() PresenceManager {
	return m.presence
}

func (m *WebSocketManager) ClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

func (m *WebSocketManager) Channels() map[string]map[*Client]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.channels
}

func (m *WebSocketManager) Stop() {
	select {
	case <-m.stopChan:

	default:
		close(m.stopChan)
	}
}
