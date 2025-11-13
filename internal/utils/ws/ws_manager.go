package ws

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager interface for WebSocket operations
type Manager interface {
	Run()
	RegisterClient(client *Client)
	UnregisterClient(client *Client)
	BroadcastToChannel(channel string, message []byte)
	SubscribeToChannel(client *Client, channel string)
	UnsubscribeFromChannel(client *Client, channel string)
	GetChannelClients(channel string) int
}

// WebSocketManager manages WebSocket connections and channels
type WebSocketManager struct {
	// Registered clients
	clients map[*Client]bool

	// Channel subscriptions: channel -> set of clients
	channels map[string]map[*Client]bool

	// Inbound messages from clients
	broadcast chan *BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Channel subscription requests
	subscribe chan *SubscriptionRequest

	// Channel unsubscription requests
	unsubscribe chan *SubscriptionRequest

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	log *logrus.Logger

	// Configuration
	config *WebSocketConfig
}

// BroadcastMessage represents a message to broadcast to a channel
type BroadcastMessage struct {
	Channel string
	Message []byte
}

// SubscriptionRequest represents a channel subscription/unsubscription request
type SubscriptionRequest struct {
	Client  *Client
	Channel string
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

// NewWebSocketManager creates a new WebSocket Manager
func NewWebSocketManager(config *WebSocketConfig, log *logrus.Logger) *WebSocketManager {
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
	}
}

// Run starts the WebSocket Manager event loop
func (m *WebSocketManager) Run() {
	m.log.Info("WebSocket Manager started")

	for {
		select {
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

// handleRegister registers a new client
func (m *WebSocketManager) handleRegister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client] = true
	m.log.Infof("Client registered: %s, total clients: %d", client.ID, len(m.clients))
}

// handleUnregister unregisters a client and removes from all channels
func (m *WebSocketManager) handleUnregister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[client]; ok {
		// Remove from all channels
		for channel, clients := range m.channels {
			if _, exists := clients[client]; exists {
				delete(clients, client)
				m.log.Infof("Client %s removed from channel: %s", client.ID, channel)

				// Clean up empty channels
				if len(clients) == 0 {
					delete(m.channels, channel)
					m.log.Infof("Channel removed (empty): %s", channel)
				}
			}
		}

		// Remove from clients map
		delete(m.clients, client)
		close(client.Send)

		m.log.Infof("Client unregistered: %s, total clients: %d", client.ID, len(m.clients))
	}
}

// handleBroadcast sends a message to all clients in a channel
func (m *WebSocketManager) handleBroadcast(msg *BroadcastMessage) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, ok := m.channels[msg.Channel]; ok {
		count := 0
		for client := range clients {
			select {
			case client.Send <- msg.Message:
				count++
			default:
				// Client's Send buffer is full, skip
				m.log.Warnf("Failed to Send message to client %s (buffer full)", client.ID)
			}
		}
		m.log.Debugf("Broadcast to channel %s: %d/%d clients", msg.Channel, count, len(clients))
	}
}

// handleSubscribe subscribes a client to a channel
func (m *WebSocketManager) handleSubscribe(req *SubscriptionRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create channel if it doesn't exist
	if _, ok := m.channels[req.Channel]; !ok {
		m.channels[req.Channel] = make(map[*Client]bool)
		m.log.Infof("Channel created: %s", req.Channel)
	}

	// Add client to channel
	m.channels[req.Channel][req.Client] = true
	m.log.Infof("Client %s subscribed to channel: %s, total subscribers: %d",
		req.Client.ID, req.Channel, len(m.channels[req.Channel]))
}

// handleUnsubscribe unsubscribes a client from a channel
func (m *WebSocketManager) handleUnsubscribe(req *SubscriptionRequest) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if clients, ok := m.channels[req.Channel]; ok {
		if _, exists := clients[req.Client]; exists {
			delete(clients, req.Client)
			m.log.Infof("Client %s unsubscribed from channel: %s, remaining subscribers: %d",
				req.Client.ID, req.Channel, len(clients))

			// Clean up empty channels
			if len(clients) == 0 {
				delete(m.channels, req.Channel)
				m.log.Infof("Channel removed (empty): %s", req.Channel)
			}
		}
	}
}

// RegisterClient registers a new client
func (m *WebSocketManager) RegisterClient(client *Client) {
	m.register <- client
}

// UnregisterClient unregisters a client
func (m *WebSocketManager) UnregisterClient(client *Client) {
	m.unregister <- client
}

// BroadcastToChannel broadcasts a message to all clients in a channel
func (m *WebSocketManager) BroadcastToChannel(channel string, message []byte) {
	m.broadcast <- &BroadcastMessage{
		Channel: channel,
		Message: message,
	}
}

// SubscribeToChannel subscribes a client to a channel
func (m *WebSocketManager) SubscribeToChannel(client *Client, channel string) {
	m.subscribe <- &SubscriptionRequest{
		Client:  client,
		Channel: channel,
	}
}

// UnsubscribeFromChannel unsubscribes a client from a channel
func (m *WebSocketManager) UnsubscribeFromChannel(client *Client, channel string) {
	m.unsubscribe <- &SubscriptionRequest{
		Client:  client,
		Channel: channel,
	}
}

// GetChannelClients returns the number of clients in a channel
func (m *WebSocketManager) GetChannelClients(channel string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if clients, ok := m.channels[channel]; ok {
		return len(clients)
	}
	return 0
}
