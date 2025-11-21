package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypePrice      MessageType = "price_update"
	MessageTypeSignal     MessageType = "signal"
	MessageTypePortfolio  MessageType = "portfolio_update"
	MessageTypeAlert      MessageType = "alert"
	MessageTypeIndicator  MessageType = "indicator"
	MessageTypeHeartbeat  MessageType = "heartbeat"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	UserID   uint
	Hub      *Hub
	Conn     *Connection
	Send     chan []byte
	Topics   map[string]bool // Subscribed topics
	mu       sync.RWMutex
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	log        *slog.Logger
}

// BroadcastMessage contains a message and optional targeting
type BroadcastMessage struct {
	Message  *Message
	UserID   uint   // 0 for all users
	Topic    string // Empty for all topics
}

// NewHub creates a new WebSocket hub
func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		log:        log,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	h.log.Info("Starting WebSocket hub")

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case broadcastMsg := <-h.broadcast:
			h.broadcastMessage(broadcastMsg)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	h.log.Info("Client registered",
		"client_id", client.ID,
		"user_id", client.UserID,
		"total_clients", len(h.clients))
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
	}
	h.mu.Unlock()

	h.log.Info("Client unregistered",
		"client_id", client.ID,
		"user_id", client.UserID,
		"total_clients", len(h.clients))
}

// broadcastMessage broadcasts a message to relevant clients
func (h *Hub) broadcastMessage(broadcastMsg *BroadcastMessage) {
	messageBytes, err := json.Marshal(broadcastMsg.Message)
	if err != nil {
		h.log.Error("Failed to marshal message", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	sentCount := 0

	for client := range h.clients {
		// Check user targeting
		if broadcastMsg.UserID != 0 && client.UserID != broadcastMsg.UserID {
			continue
		}

		// Check topic subscription
		if broadcastMsg.Topic != "" {
			client.mu.RLock()
			subscribed := client.Topics[broadcastMsg.Topic]
			client.mu.RUnlock()

			if !subscribed {
				continue
			}
		}

		// Send message
		select {
		case client.Send <- messageBytes:
			sentCount++
		default:
			// Client's send buffer is full, unregister it
			go func(c *Client) {
				h.Unregister <- c
			}(client)
		}
	}

	if sentCount > 0 {
		h.log.Debug("Message broadcasted",
			"type", broadcastMsg.Message.Type,
			"recipients", sentCount)
	}
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(msgType MessageType, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		h.log.Error("Failed to marshal data", "error", err)
		return
	}

	message := &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}

	h.broadcast <- &BroadcastMessage{
		Message: message,
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID uint, msgType MessageType, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		h.log.Error("Failed to marshal data", "error", err)
		return
	}

	message := &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}

	h.broadcast <- &BroadcastMessage{
		Message: message,
		UserID:  userID,
	}
}

// BroadcastToTopic sends a message to clients subscribed to a topic
func (h *Hub) BroadcastToTopic(topic string, msgType MessageType, data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		h.log.Error("Failed to marshal data", "error", err)
		return
	}

	message := &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}

	h.broadcast <- &BroadcastMessage{
		Message: message,
		Topic:   topic,
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetUserClientCount returns the number of clients for a specific user
func (h *Hub) GetUserClientCount(userID uint) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for client := range h.clients {
		if client.UserID == userID {
			count++
		}
	}

	return count
}

// NewClient creates a new client
func NewClient(id string, userID uint, hub *Hub, conn *Connection) *Client {
	return &Client{
		ID:     id,
		UserID: userID,
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Topics: make(map[string]bool),
	}
}

// Subscribe subscribes the client to a topic
func (c *Client) Subscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Topics[topic] = true
	c.Hub.log.Debug("Client subscribed to topic",
		"client_id", c.ID,
		"topic", topic)
}

// Unsubscribe unsubscribes the client from a topic
func (c *Client) Unsubscribe(topic string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.Topics, topic)
	c.Hub.log.Debug("Client unsubscribed from topic",
		"client_id", c.ID,
		"topic", topic)
}

// IsSubscribed checks if client is subscribed to a topic
func (c *Client) IsSubscribed(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Topics[topic]
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
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
			if !IsNormalClose(err) {
				c.Hub.log.Error("Read error", "client_id", c.ID, "error", err)
			}
			break
		}

		// Handle incoming messages (subscriptions, etc.)
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection
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
				// Hub closed the channel
				c.Conn.WriteMessage(CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(TextMessage, message); err != nil {
				c.Hub.log.Error("Write error", "client_id", c.ID, "error", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from clients
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Action string `json:"action"`
		Topic  string `json:"topic"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		c.Hub.log.Error("Failed to parse message", "error", err)
		return
	}

	switch msg.Action {
	case "subscribe":
		c.Subscribe(msg.Topic)
	case "unsubscribe":
		c.Unsubscribe(msg.Topic)
	default:
		c.Hub.log.Warn("Unknown action", "action", msg.Action)
	}
}
