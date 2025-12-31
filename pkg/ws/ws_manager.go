package ws

import (
	"context"
	"sync"
	"time"

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
}

type BroadcastMessage struct {
	Channel    string
	Message    []byte
	FromRemote bool // Internal flag to avoid re-publishing to Redis
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

// NewWebSocketManager creates a new WebSocketManager with the provided configuration, logger, and redis client.
func NewWebSocketManager(config *WebSocketConfig, log *logrus.Logger, redisClient *redis.Client) *WebSocketManager {
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
	}
}

func (m *WebSocketManager) Run() {
	m.log.Info("WebSocket Manager started")

	// Start Redis subscriber if enabled and redis is available
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

// listenToRedis listens for broadcast messages from other instances via Redis Pub/Sub
func (m *WebSocketManager) listenToRedis() {
	ctx := context.Background()
	prefix := m.config.RedisPrefix
	if prefix == "" {
		prefix = "ws_broadcast:"
	}

	// Use pattern-based subscribe to listen to all ws_broadcast channels
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
			// Extract local channel name from Redis channel name (strip prefix)
			localChannel := msg.Channel[len(prefix):]

			// Inject into local broadcast queue with FromRemote = true
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
	m.log.Infof("Client registered: %s, total clients: %d", client.ID, len(m.clients))
}

func (m *WebSocketManager) handleUnregister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[client]; ok {

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

		m.log.Infof("Client unregistered: %s, total clients: %d", client.ID, len(m.clients))
	}
}

func (m *WebSocketManager) handleBroadcast(msg *BroadcastMessage) {
	// 1. Publish to Redis if this is a local broadcast, feature is ENABLED, and redis is available
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
	}

	// 2. Send to local clients
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, ok := m.channels[msg.Channel]; ok {
		count := 0
		for client := range clients {
			select {
			case client.Send <- msg.Message:
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

// RegisterClient registers a client with a timeout
func (m *WebSocketManager) RegisterClient(client *Client) {
	select {
	case m.register <- client:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("RegisterClient timed out")
	case <-m.stopChan:
		m.log.Warn("RegisterClient called on stopped manager")
	}
}

// UnregisterClient unregisters a client with a timeout
func (m *WebSocketManager) UnregisterClient(client *Client) {
	select {
	case m.unregister <- client:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("UnregisterClient timed out")
	case <-m.stopChan:
		m.log.Warn("UnregisterClient called on stopped manager")
	}
}

// BroadcastToChannel sends a message to a channel with a timeout
func (m *WebSocketManager) BroadcastToChannel(channel string, message []byte) {
	select {
	case m.broadcast <- &BroadcastMessage{Channel: channel, Message: message, FromRemote: false}:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("BroadcastToChannel timed out")
	case <-m.stopChan:
		m.log.Warn("BroadcastToChannel called on stopped manager")
	}
}

// SubscribeToChannel subscribes a client to a channel with a timeout
func (m *WebSocketManager) SubscribeToChannel(client *Client, channel string) {
	select {
	case m.subscribe <- &SubscriptionRequest{Client: client, Channel: channel}:
	case <-time.After(100 * time.Millisecond):
		m.log.Warn("SubscribeToChannel timed out")
	case <-m.stopChan:
		m.log.Warn("SubscribeToChannel called on stopped manager")
	}
}

// UnsubscribeFromChannel unsubscribes a client from a channel with a timeout
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
		// Already closed
	default:
		close(m.stopChan)
	}
}
