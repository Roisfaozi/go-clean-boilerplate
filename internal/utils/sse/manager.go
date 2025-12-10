package sse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

// Event represents the data structure of an SSE event
type Event struct {
	Name string      `json:"name"` // Event name (e.g., "message", "update")
	Data interface{} `json:"data"` // Payload
}

// Client represents a single connection
type Client struct {
	Channel chan Event
}

// Manager handles the connected clients and broadcasting
type Manager struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	mutex      sync.Mutex
}

// NewManager creates a new SSE Manager instance
func NewManager() *Manager {
	m := &Manager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event),
	}
	go m.run()
	return m
}

// run is the main loop that handles channels safely
func (m *Manager) run() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client] = true
			m.mutex.Unlock()
			log.Println("SSE: New client connected")

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.Channel)
			}
			m.mutex.Unlock()
			log.Println("SSE: Client disconnected")

		case event := <-m.broadcast:
			m.mutex.Lock()
			for client := range m.clients {
				select {
				case client.Channel <- event:
				default:
					// If channel is blocked/full, remove client to prevent deadlock
					delete(m.clients, client)
					close(client.Channel)
				}
			}
			m.mutex.Unlock()
		}
	}
}

// Broadcast sends an event to all connected clients
func (m *Manager) Broadcast(eventName string, data interface{}) {
	m.broadcast <- Event{
		Name: eventName,
		Data: data,
	}
}

// ServeHTTP is the Gin handler to stream events to the client
func (m *Manager) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set headers for SSE
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// Create a new client channel
		clientChan := make(chan Event)
		client := &Client{Channel: clientChan}

		// Register client
		m.register <- client

		// Listen for connection close
		// Ideally use c.Request.Context().Done()
		defer func() {
			m.unregister <- client
		}()

		// Stream events
		c.Stream(func(w io.Writer) bool {
			select {
			case <-c.Request.Context().Done():
				return false // Client disconnected
			case event, ok := <-clientChan:
				if !ok {
					return false // Channel closed
				}
				// Format: 
			// event: <event_name>\n
				// data: <json_data>\n\n
				c.Writer.Write([]byte(fmt.Sprintf("event: %s\n", event.Name)))
				
				jsonData, err := json.Marshal(event.Data)
				if err != nil {
					// Fallback for string
					c.Writer.Write([]byte(fmt.Sprintf("data: %v\n\n", event.Data)))
				} else {
					c.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData)))
				}
				
				c.Writer.Flush()
				return true
			}
		})
	}
}
