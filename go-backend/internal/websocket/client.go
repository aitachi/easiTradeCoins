package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID *uint) *Client {
	return &Client{
		ID:            uuid.New().String(),
		Hub:           hub,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
		UserID:        userID,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle message
		c.handleMessage(&msg)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg)
	case "unsubscribe":
		c.handleUnsubscribe(msg)
	case "ping":
		c.handlePing()
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleSubscribe handles subscribe requests
func (c *Client) handleSubscribe(msg *Message) {
	if msg.Channel == "" {
		c.sendError("Channel is required for subscription")
		return
	}

	symbol := msg.Symbol
	if symbol == "" {
		symbol = "*" // Subscribe to all symbols
	}

	c.Hub.Subscribe(c, msg.Channel, symbol)

	// Send confirmation
	response := Message{
		Type:    "subscribed",
		Channel: msg.Channel,
		Symbol:  symbol,
	}
	c.sendMessage(&response)
}

// handleUnsubscribe handles unsubscribe requests
func (c *Client) handleUnsubscribe(msg *Message) {
	if msg.Channel == "" {
		c.sendError("Channel is required for unsubscription")
		return
	}

	symbol := msg.Symbol
	if symbol == "" {
		symbol = "*"
	}

	c.Hub.Unsubscribe(c, msg.Channel, symbol)

	// Send confirmation
	response := Message{
		Type:    "unsubscribed",
		Channel: msg.Channel,
		Symbol:  symbol,
	}
	c.sendMessage(&response)
}

// handlePing handles ping requests
func (c *Client) handlePing() {
	response := Message{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}
	c.sendMessage(&response)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		log.Printf("Client %s send channel is full", c.ID)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(errMsg string) {
	msg := Message{
		Type: "error",
		Data: map[string]interface{}{
			"error": errMsg,
		},
	}
	c.sendMessage(&msg)
}

// GetSubscriptions returns the client's current subscriptions
func (c *Client) GetSubscriptions() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	subs := make([]string, 0, len(c.Subscriptions))
	for sub := range c.Subscriptions {
		subs = append(subs, sub)
	}
	return subs
}
