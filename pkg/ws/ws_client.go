package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Client struct {
	ID      string
	Conn    *websocket.Conn
	Manager Manager
	Send    chan []byte
	Log     *logrus.Logger
	Config  *WebSocketConfig
}

type ClientMessage struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type ServerMessage struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

// NewWebsocketClient creates a new WebSocket client
//
// conn: The WebSocket connection to the client
// manager: The WebSocket manager to handle client events
// log: The logger to log client events
// config: The WebSocket configuration options
//
// Returns a pointer to the newly created client
func NewWebsocketClient(conn *websocket.Conn, manager Manager, log *logrus.Logger, config *WebSocketConfig) *Client {
	id, err := uuid.NewV7()
	if err != nil {
		log.Errorf("Failed to generate UUID v7 for websocket client: %v", err)
		panic(err)
	}

	return &Client{
		ID:      id.String(),
		Conn:    conn,
		Manager: manager,
		Send:    make(chan []byte, 256),
		Log:     log,
		Config:  config,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(c.Config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Config.PongWait))
		return nil
	})

	c.Conn.SetReadLimit(c.Config.MaxMessageSize)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Log.Errorf("WebSocket error for client %s: %v", c.ID, err)
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.Log.Errorf("Failed to unmarshal message from client %s: %v", c.ID, err)
		c.sendError("Invalid message format")
		return
	}

	switch msg.Type {
	case "subscribe":
		if msg.Channel == "" {
			c.sendError("Channel name is required")
			return
		}
		c.Manager.SubscribeToChannel(c, msg.Channel)
		c.sendInfo(msg.Channel, "Subscribed to channel")

	case "unsubscribe":
		if msg.Channel == "" {
			c.sendError("Channel name is required")
			return
		}
		c.Manager.UnsubscribeFromChannel(c, msg.Channel)
		c.sendInfo(msg.Channel, "Unsubscribed from channel")

	case "message":
		if msg.Channel == "" {
			c.sendError("Channel name is required")
			return
		}
		serverMsg := ServerMessage{
			Type:    "message",
			Channel: msg.Channel,
			Data:    msg.Data,
		}
		msgBytes, err := json.Marshal(serverMsg)
		if err != nil {
			c.Log.Errorf("Failed to marshal message: %v", err)
			return
		}
		c.Manager.BroadcastToChannel(msg.Channel, msgBytes)

	default:
		c.sendError("Unknown message type")
	}
}

func (c *Client) sendError(errorMsg string) {
	msg := ServerMessage{
		Type: "error",
		Data: errorMsg,
	}
	c.sendMessage(msg)
}

func (c *Client) sendInfo(channel, infoMsg string) {
	msg := ServerMessage{
		Type:    "info",
		Channel: channel,
		Data:    infoMsg,
	}
	c.sendMessage(msg)
}

func (c *Client) sendMessage(msg ServerMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		c.Log.Errorf("Failed to marshal message: %v", err)
		return
	}

	select {
	case c.Send <- msgBytes:
	default:
		c.Log.Warnf("Client %s Send buffer is full", c.ID)
	}
}
