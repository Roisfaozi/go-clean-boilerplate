package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type Client struct {
	ID      string
	Manager Manager
	Conn    *websocket.Conn
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
	Channel string      `json:"channel,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// NewWebsocketClient creates a new Client instance
func NewWebsocketClient(conn *websocket.Conn, manager Manager, log *logrus.Logger, config *WebSocketConfig) *Client {
	uid, err := uuid.NewV7()
	if err != nil {
		log.Errorf("Failed to generate UUID for client: %v", err)
		return nil // Or handle error appropriately
	}
	return &Client{
		ID:      uid.String(),
		Manager: manager,
		Conn:    conn,
		Send:    make(chan []byte, 256), // Buffered channel
		Log:     log,
		Config:  config,
	}
}

// ReadPump pumps messages from the websocket connection to the Manager.
//
// The application ensures that there is at most one reader per websocket connection
// running at any given time.
func (c *Client) ReadPump() {
	defer func() {
		c.Manager.UnregisterClient(c)
		if err := c.Conn.Close(); err != nil {
			c.Log.Warnf("Error closing connection for client %s: %v", c.ID, err)
		}
	}()
	c.Conn.SetReadLimit(c.Config.MaxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.Config.PongWait)); err != nil {
		c.Log.Errorf("Client %s: SetReadDeadline failed: %v", c.ID, err)
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(c.Config.PongWait)); err != nil {
			c.Log.Errorf("Client %s: SetReadDeadline in pong handler failed: %v", c.ID, err)
			return err
		}
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Log.Errorf("Client %s: unexpected close error: %v", c.ID, err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the Manager to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer per websocket
// connection running at any given time.
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			c.Log.Warnf("Error closing connection for client %s: %v", c.ID, err)
		}
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait)); err != nil {
				c.Log.Errorf("Client %s: SetWriteDeadline failed: %v", c.ID, err)
				return
			}
			if !ok {
				// The Manager closed the Send channel.
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					c.Log.Errorf("Client %s: Write CloseMessage failed: %v", c.ID, err)
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.Log.Errorf("Client %s: NextWriter failed: %v", c.ID, err)
				return
			}
			if _, err := w.Write(message); err != nil {
				c.Log.Errorf("Client %s: write message failed: %v", c.ID, err)
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					c.Log.Errorf("Client %s: write newline failed: %v", c.ID, err)
					return
				}
				if _, err := w.Write(<-c.Send); err != nil {
					c.Log.Errorf("Client %s: write queued message failed: %v", c.ID, err)
					return
				}
			}

			if err := w.Close(); err != nil {
				c.Log.Errorf("Client %s: writer close failed: %v", c.ID, err)
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait)); err != nil {
				c.Log.Errorf("Client %s: SetWriteDeadline on ping failed: %v", c.ID, err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Log.Errorf("Client %s: Write PingMessage failed: %v", c.ID, err)
				return
			}
		}
	}
}

func (c *Client) handleMessage(message []byte) {
	var clientMsg ClientMessage
	if err := json.Unmarshal(message, &clientMsg); err != nil {
		c.Log.Warnf("Client %s: Failed to unmarshal client message: %v", c.ID, err)
		c.sendError("Invalid message format")
		return
	}

	switch clientMsg.Type {
	case "subscribe":
		c.Manager.SubscribeToChannel(c, clientMsg.Channel)
		c.sendInfo(clientMsg.Channel, fmt.Sprintf("Subscribed to channel: %s", clientMsg.Channel))
	case "unsubscribe":
		c.Manager.UnsubscribeFromChannel(c, clientMsg.Channel)
		c.sendInfo(clientMsg.Channel, fmt.Sprintf("Unsubscribed from channel: %s", clientMsg.Channel))
	case "message":
		c.Log.Infof("Client %s sent message to channel %s: %s", c.ID, clientMsg.Channel, clientMsg.Data)
		// For now, we only handle subscribe/unsubscribe/info. Actual message broadcast is handled by manager.
		// If client sends "message" type, we can broadcast it here directly or via manager.
		// c.Manager.BroadcastToChannel(clientMsg.Channel, message)
	default:
		c.Log.Warnf("Client %s: Unknown message type: %s", c.ID, clientMsg.Type)
		c.sendError(fmt.Sprintf("Unknown message type: %s", clientMsg.Type))
	}
}

func (c *Client) sendInfo(channel, data string) {
	msg := ServerMessage{
		Type:    "info",
		Channel: channel,
		Data:    data,
	}
	c.sendJSON(msg)
}

func (c *Client) sendError(data string) {
	msg := ServerMessage{
		Type: "error",
		Data: data,
	}
	c.sendJSON(msg)
}

func (c *Client) sendJSON(data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		c.Log.Errorf("Client %s: Failed to marshal JSON for sending: %v", c.ID, err)
		return
	}
	select {
	case c.Send <- payload:
	default:
		c.Log.Warnf("Client %s: Send buffer full, dropping message", c.ID)
	}
}
