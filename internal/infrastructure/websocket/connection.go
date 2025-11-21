package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket message types
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
	CloseMessage  = websocket.CloseMessage
	PingMessage   = websocket.PingMessage
	PongMessage   = websocket.PongMessage
)

// Connection wraps a gorilla websocket connection
type Connection struct {
	ws *websocket.Conn
}

// NewConnection creates a new connection wrapper
func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{ws: ws}
}

// ReadMessage reads a message from the websocket
func (c *Connection) ReadMessage() (messageType int, p []byte, err error) {
	return c.ws.ReadMessage()
}

// WriteMessage writes a message to the websocket
func (c *Connection) WriteMessage(messageType int, data []byte) error {
	return c.ws.WriteMessage(messageType, data)
}

// Close closes the websocket connection
func (c *Connection) Close() error {
	return c.ws.Close()
}

// SetReadDeadline sets the read deadline
func (c *Connection) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (c *Connection) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}

// SetPongHandler sets the pong handler
func (c *Connection) SetPongHandler(h func(appData string) error) {
	c.ws.SetPongHandler(h)
}

// IsNormalClose checks if error is a normal close error
func IsNormalClose(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}
