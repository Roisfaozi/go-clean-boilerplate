package ws

import (
	"sync"
	"time"

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
}

type BroadcastMessage struct {
	Channel string
	Message []byte
}

type SubscriptionRequest struct {
	Client  *Client
	Channel string
}

type WebSocketConfig struct {
	WriteWait      time.Duration
	PongWait       time.Duration
	PingPeriod     time.Duration
	MaxMessageSize int64
}

// NewWebSocketManager creates a new WebSocketManager with the provided configuration and logger.
//
// config: The WebSocket configuration options.
// log: The logger to log manager events.
//
// Returns a pointer to the newly created WebSocketManager.
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
		m.log.Debugf("Broadcast to channel %s: %d/%d clients", msg.Channel, count, len(clients))
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

func (m *WebSocketManager) RegisterClient(client *Client) {
	m.register <- client
}

func (m *WebSocketManager) UnregisterClient(client *Client) {
	m.unregister <- client
}

func (m *WebSocketManager) BroadcastToChannel(channel string, message []byte) {
	m.broadcast <- &BroadcastMessage{
		Channel: channel,
		Message: message,
	}
}

func (m *WebSocketManager) SubscribeToChannel(client *Client, channel string) {
	m.subscribe <- &SubscriptionRequest{
		Client:  client,
		Channel: channel,
	}
}

func (m *WebSocketManager) UnsubscribeFromChannel(client *Client, channel string) {
	m.unsubscribe <- &SubscriptionRequest{
		Client:  client,
		Channel: channel,
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
