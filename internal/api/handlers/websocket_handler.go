package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"go-finance-advisor/internal/infrastructure/middleware"
	ws "go-finance-advisor/internal/infrastructure/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *ws.Hub
	log *slog.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *ws.Hub, log *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		log: log,
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Error("Failed to upgrade connection", "error", err)
		return
	}

	// Create client
	clientID := uuid.New().String()
	connection := ws.NewConnection(conn)
	client := ws.NewClient(clientID, userID, h.hub, connection)

	// Register client with hub
	h.hub.Register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()

	h.log.Info("WebSocket connection established",
		"client_id", clientID,
		"user_id", userID)
}

// HandleHealth returns WebSocket hub health status
func (h *WebSocketHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":         "healthy",
		"client_count":   h.hub.GetClientCount(),
	})
}

// RegisterRoutes registers WebSocket routes
func (h *WebSocketHandler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	wsGroup := router.Group("/ws")
	{
		// WebSocket endpoint (requires authentication)
		wsGroup.GET("/connect", authMiddleware.Authenticate(), h.HandleWebSocket)

		// Health check
		wsGroup.GET("/health", h.HandleHealth)
	}

	h.log.Info("WebSocket routes registered")
}

// BroadcastPriceUpdate broadcasts a price update to all clients
func (h *WebSocketHandler) BroadcastPriceUpdate(data interface{}) {
	h.hub.Broadcast(ws.MessageTypePrice, data)
}

// BroadcastSignal broadcasts a new signal to all clients
func (h *WebSocketHandler) BroadcastSignal(data interface{}) {
	h.hub.Broadcast(ws.MessageTypeSignal, data)
}

// BroadcastPortfolioUpdate broadcasts portfolio update to a specific user
func (h *WebSocketHandler) BroadcastPortfolioUpdate(userID uint, data interface{}) {
	h.hub.BroadcastToUser(userID, ws.MessageTypePortfolio, data)
}

// BroadcastAlert broadcasts an alert to a specific user
func (h *WebSocketHandler) BroadcastAlert(userID uint, data interface{}) {
	h.hub.BroadcastToUser(userID, ws.MessageTypeAlert, data)
}

// BroadcastToSymbol broadcasts to clients watching a specific symbol
func (h *WebSocketHandler) BroadcastToSymbol(symbol string, msgType ws.MessageType, data interface{}) {
	topic := fmt.Sprintf("symbol:%s", symbol)
	h.hub.BroadcastToTopic(topic, msgType, data)
}
