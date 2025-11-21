package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	binanceWSEndpoint = "wss://stream.binance.com:9443/ws"
	pingInterval      = 3 * time.Minute
	reconnectDelay    = 5 * time.Second
	maxReconnectDelay = 5 * time.Minute
)

// TickerData represents real-time ticker data from Binance
type TickerData struct {
	Symbol             string    `json:"s"`
	PriceChange        string    `json:"p"`
	PriceChangePercent string    `json:"P"`
	LastPrice          string    `json:"c"`
	Volume             string    `json:"v"`
	QuoteVolume        string    `json:"q"`
	HighPrice          string    `json:"h"`
	LowPrice           string    `json:"l"`
	OpenPrice          string    `json:"o"`
	EventTime          int64     `json:"E"`
	Timestamp          time.Time `json:"-"`
}

// TradeData represents real-time trade data
type TradeData struct {
	Symbol    string  `json:"s"`
	Price     float64 `json:"p,string"`
	Quantity  float64 `json:"q,string"`
	EventTime int64   `json:"E"`
	TradeTime int64   `json:"T"`
	IsBuyerMaker bool `json:"m"`
}

// KlineData represents candlestick data
type KlineData struct {
	Symbol    string `json:"s"`
	EventTime int64  `json:"E"`
	Kline     struct {
		StartTime            int64   `json:"t"`
		EndTime              int64   `json:"T"`
		Symbol               string  `json:"s"`
		Interval             string  `json:"i"`
		Open                 float64 `json:"o,string"`
		Close                float64 `json:"c,string"`
		High                 float64 `json:"h,string"`
		Low                  float64 `json:"l,string"`
		Volume               float64 `json:"v,string"`
		IsClosed             bool    `json:"x"`
		QuoteAssetVolume     float64 `json:"q,string"`
		NumberOfTrades       int64   `json:"n"`
		TakerBuyBaseVolume   float64 `json:"V,string"`
		TakerBuyQuoteVolume  float64 `json:"Q,string"`
	} `json:"k"`
}

// WebSocketMessage represents a generic WebSocket message
type WebSocketMessage struct {
	EventType string          `json:"e"`
	EventTime int64           `json:"E"`
	Symbol    string          `json:"s"`
	Data      json.RawMessage `json:",inline"`
}

// DataHandler is called when new market data is received
type DataHandler func(data interface{})

// WebSocketClient manages Binance WebSocket connections
type WebSocketClient struct {
	conn             *websocket.Conn
	mu               sync.RWMutex
	log              *slog.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	subscriptions    []string
	reconnectDelay   time.Duration
	onTicker         DataHandler
	onTrade          DataHandler
	onKline          DataHandler
	isConnected      bool
	reconnectBackoff time.Duration
}

// NewWebSocketClient creates a new Binance WebSocket client
func NewWebSocketClient(log *slog.Logger) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketClient{
		log:              log,
		ctx:              ctx,
		cancel:           cancel,
		reconnectDelay:   reconnectDelay,
		reconnectBackoff: reconnectDelay,
		subscriptions:    make([]string, 0),
	}
}

// OnTicker sets the handler for ticker data
func (c *WebSocketClient) OnTicker(handler DataHandler) {
	c.onTicker = handler
}

// OnTrade sets the handler for trade data
func (c *WebSocketClient) OnTrade(handler DataHandler) {
	c.onTrade = handler
}

// OnKline sets the handler for kline/candlestick data
func (c *WebSocketClient) OnKline(handler DataHandler) {
	c.onKline = handler
}

// Connect establishes WebSocket connection
func (c *WebSocketClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return nil
	}

	c.log.Info("Connecting to Binance WebSocket", "endpoint", binanceWSEndpoint)

	conn, _, err := websocket.DefaultDialer.Dial(binanceWSEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}

	c.conn = conn
	c.isConnected = true
	c.reconnectBackoff = reconnectDelay

	c.log.Info("Connected to Binance WebSocket successfully")

	// Start ping routine
	go c.pingRoutine()

	// Start read routine
	go c.readRoutine()

	// Resubscribe if we have previous subscriptions
	if len(c.subscriptions) > 0 {
		c.log.Info("Resubscribing to streams", "count", len(c.subscriptions))
		if err := c.subscribe(c.subscriptions); err != nil {
			c.log.Error("Failed to resubscribe", "error", err)
		}
	}

	return nil
}

// subscribe sends subscription request to WebSocket
func (c *WebSocketClient) subscribe(streams []string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	subscribeMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": streams,
		"id":     time.Now().Unix(),
	}

	if err := c.conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	c.log.Info("Subscribed to streams", "streams", strings.Join(streams, ", "))
	return nil
}

// SubscribeTicker subscribes to 24hr ticker for symbols
func (c *WebSocketClient) SubscribeTicker(symbols []string) error {
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = fmt.Sprintf("%s@ticker", strings.ToLower(symbol))
	}

	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, streams...)
	c.mu.Unlock()

	if c.isConnected {
		return c.subscribe(streams)
	}

	return nil
}

// SubscribeTrades subscribes to trade streams for symbols
func (c *WebSocketClient) SubscribeTrades(symbols []string) error {
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	}

	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, streams...)
	c.mu.Unlock()

	if c.isConnected {
		return c.subscribe(streams)
	}

	return nil
}

// SubscribeKlines subscribes to kline/candlestick streams
func (c *WebSocketClient) SubscribeKlines(symbols []string, interval string) error {
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
	}

	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, streams...)
	c.mu.Unlock()

	if c.isConnected {
		return c.subscribe(streams)
	}

	return nil
}

// pingRoutine sends periodic ping messages to keep connection alive
func (c *WebSocketClient) pingRoutine() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			if c.conn != nil {
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.log.Error("Failed to send ping", "error", err)
					c.isConnected = false
				}
			}
			c.mu.Unlock()
		}
	}
}

// readRoutine reads messages from WebSocket
func (c *WebSocketClient) readRoutine() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				c.log.Error("WebSocket read error", "error", err)
				c.mu.Lock()
				c.isConnected = false
				c.mu.Unlock()
				c.reconnect()
				return
			}

			c.handleMessage(message)
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *WebSocketClient) handleMessage(message []byte) {
	var wsMsg map[string]interface{}
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		c.log.Error("Failed to unmarshal message", "error", err)
		return
	}

	eventType, ok := wsMsg["e"].(string)
	if !ok {
		// Not an event message (could be subscription confirmation)
		return
	}

	switch eventType {
	case "24hrTicker":
		if c.onTicker != nil {
			var ticker TickerData
			if err := json.Unmarshal(message, &ticker); err != nil {
				c.log.Error("Failed to parse ticker data", "error", err)
				return
			}
			ticker.Timestamp = time.UnixMilli(ticker.EventTime)
			c.onTicker(ticker)
		}

	case "trade":
		if c.onTrade != nil {
			var trade TradeData
			if err := json.Unmarshal(message, &trade); err != nil {
				c.log.Error("Failed to parse trade data", "error", err)
				return
			}
			c.onTrade(trade)
		}

	case "kline":
		if c.onKline != nil {
			var kline KlineData
			if err := json.Unmarshal(message, &kline); err != nil {
				c.log.Error("Failed to parse kline data", "error", err)
				return
			}
			c.onKline(kline)
		}
	}
}

// reconnect attempts to reconnect with exponential backoff
func (c *WebSocketClient) reconnect() {
	c.log.Info("Attempting to reconnect", "delay", c.reconnectBackoff)

	time.Sleep(c.reconnectBackoff)

	if err := c.Connect(); err != nil {
		c.log.Error("Reconnection failed", "error", err)

		// Exponential backoff
		c.reconnectBackoff *= 2
		if c.reconnectBackoff > maxReconnectDelay {
			c.reconnectBackoff = maxReconnectDelay
		}

		go c.reconnect()
		return
	}

	c.log.Info("Reconnected successfully")
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	c.cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.isConnected = false
		return c.conn.Close()
	}

	return nil
}

// IsConnected returns the connection status
func (c *WebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}
