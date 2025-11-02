//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/easitradecoins/backend/internal/models"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	ID            string
	Conn          *websocket.Conn
	Send          chan []byte
	Hub           *Hub
	Subscriptions map[string]bool // channel -> subscribed
	mu            sync.RWMutex
}

// Hub manages WebSocket clients and broadcasts
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 10000),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s, total clients: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s, total clients: %d", client.ID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToChannel broadcasts a message to a specific channel
func (h *Hub) BroadcastToChannel(channel string, data interface{}) {
	message := Message{
		Type:    "update",
		Channel: channel,
		Data:    data,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.IsSubscribed(channel) {
			select {
			case client.Send <- jsonData:
			default:
				// Client send buffer is full, skip
			}
		}
	}
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SubscribeMessage represents a subscribe message
type SubscribeMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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

		// Handle subscribe/unsubscribe messages
		var subMsg SubscribeMessage
		if err := json.Unmarshal(message, &subMsg); err == nil {
			c.handleSubscription(&subMsg)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
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
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscription handles subscription messages
func (c *Client) handleSubscription(msg *SubscribeMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch msg.Method {
	case "SUBSCRIBE":
		for _, channel := range msg.Params {
			c.Subscriptions[channel] = true
		}
		// Send confirmation
		response := Message{
			Type: "subscribed",
			Data: msg.Params,
		}
		c.sendMessage(response)

	case "UNSUBSCRIBE":
		for _, channel := range msg.Params {
			delete(c.Subscriptions, channel)
		}
		// Send confirmation
		response := Message{
			Type: "unsubscribed",
			Data: msg.Params,
		}
		c.sendMessage(response)
	}
}

// IsSubscribed checks if client is subscribed to a channel
func (c *Client) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscriptions[channel]
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.Send <- data:
	default:
	}
}

// BroadcastTrade broadcasts a trade to subscribers
func (h *Hub) BroadcastTrade(trade *models.Trade) {
	channel := trade.Symbol + "@trade"
	h.BroadcastToChannel(channel, map[string]interface{}{
		"e": "trade",
		"s": trade.Symbol,
		"t": trade.TradeTime.Unix(),
		"p": trade.Price.String(),
		"q": trade.Quantity.String(),
		"m": trade.BuyOrderID < trade.SellOrderID,
	})
}

// BroadcastOrderBookUpdate broadcasts order book update
func (h *Hub) BroadcastOrderBookUpdate(symbol string, bids, asks interface{}) {
	channel := symbol + "@depth"
	h.BroadcastToChannel(channel, map[string]interface{}{
		"e": "depthUpdate",
		"s": symbol,
		"b": bids,
		"a": asks,
	})
}
