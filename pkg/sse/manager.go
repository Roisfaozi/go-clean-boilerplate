package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	_ "github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Event struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Client struct {
	Channel chan Event
}

type Manager struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mutex      sync.Mutex
	log        *logrus.Logger
	stopChan   chan struct{}
}

func NewManager() *Manager {
	m := &Manager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event),
		log:        logrus.New(),
		stopChan:   make(chan struct{}),
	}
	m.log.SetOutput(io.Discard)
	go m.run()
	return m
}

func (m *Manager) SetLogger(logger *logrus.Logger) {
	m.log = logger
}

func (m *Manager) run() {
	for {
		select {
		case <-m.stopChan:
			return
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client] = true
			m.mutex.Unlock()
			m.log.Println("SSE: New client connected")

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.Channel)
			}
			m.mutex.Unlock()
			m.log.Println("SSE: Client disconnected")

		case event := <-m.broadcast:
			m.mutex.Lock()
			for client := range m.clients {
				select {
				case client.Channel <- event:
				default:

					delete(m.clients, client)
					close(client.Channel)
				}
			}
			m.mutex.Unlock()
		}
	}
}

func (m *Manager) Broadcast(eventName string, data interface{}) {
	m.broadcast <- Event{
		Name: eventName,
		Data: data,
	}
}

func (m *Manager) ClientCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.clients)
}

func (m *Manager) Stop() {
	close(m.stopChan)
}

func (m *Manager) RegisterClient(client *Client) {
	m.register <- client
}

func (m *Manager) UnregisterClient(client *Client) {
	m.unregister <- client
}

// ServeHTTP returns standard net/http handler.
func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher.Flush()

	clientChan := make(chan Event, 10)
	client := &Client{Channel: clientChan}
	m.register <- client
	defer func() {
		m.unregister <- client
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-clientChan:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(w, "event: %s\n", event.Name); err != nil {
				return
			}
			jsonData, err := json.Marshal(event.Data)
			if err != nil {
				if _, err := fmt.Fprintf(w, "data: %v\n\n", event.Data); err != nil {
					return
				}
			} else {
				if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
					return
				}
			}
			flusher.Flush()
		}
	}
}

// GinServeHTTP returns gin-compatible handler wrapper.
func (m *Manager) GinServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.ServeHTTP(c.Writer, c.Request)
	}
}
