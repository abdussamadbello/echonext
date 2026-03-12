// Package websocket provides type-safe WebSocket support with automatic OpenAPI documentation
package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

// WebSocket configuration defaults
const (
	DefaultReadBufferSize  = 1024
	DefaultWriteBufferSize = 1024
	DefaultPingInterval    = 30 * time.Second
	DefaultPongTimeout     = 60 * time.Second
	DefaultMaxMessageSize  = 512 * 1024 // 512KB
)

// Config configures WebSocket behavior
type Config struct {
	// ReadBufferSize specifies the I/O buffer size in bytes for reads
	ReadBufferSize int
	// WriteBufferSize specifies the I/O buffer size in bytes for writes
	WriteBufferSize int
	// PingInterval is the interval between ping messages
	PingInterval time.Duration
	// PongTimeout is the timeout for pong response
	PongTimeout time.Duration
	// MaxMessageSize is the maximum message size in bytes
	MaxMessageSize int64
	// CheckOrigin is a function to validate the request origin
	CheckOrigin func(r *http.Request) bool
}

// Route configures WebSocket route metadata for OpenAPI documentation
type Route struct {
	Summary     string
	Description string
	Tags        []string
	Protocols   []string // Sub-protocols supported
}

// Connection wraps a WebSocket connection with type-safe operations
type Connection struct {
	// ID is a unique identifier for this connection
	ID string
	// conn is the underlying WebSocket connection
	conn *websocket.Conn
	// context is the Echo context
	context echo.Context
	// metadata stores custom data associated with the connection
	metadata map[string]interface{}
	// mu protects metadata
	mu sync.RWMutex
	// sendChan is used for writing messages
	sendChan chan []byte
	// done signals connection closure
	done chan struct{}
	// closed indicates if connection is closed
	closed bool
	// closeMu protects closed flag
	closeMu sync.Mutex
}

// Message wraps messages with metadata
type Message[T any] struct {
	Type      string    `json:"type"`
	Payload   T         `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	ID        string    `json:"id,omitempty"`
}

// Handler is the interface for WebSocket handlers with lifecycle methods
type Handler interface {
	// OnConnect is called when a client connects
	OnConnect(conn *Connection) error
	// OnMessage is called when a message is received
	OnMessage(conn *Connection, messageType int, data []byte) error
	// OnDisconnect is called when a client disconnects
	OnDisconnect(conn *Connection, err error)
}

// Hub manages multiple WebSocket connections
type Hub struct {
	// connections maps connection IDs to connections
	connections map[string]*Connection
	// register channel for new connections
	register chan *Connection
	// unregister channel for closing connections
	unregister chan *Connection
	// broadcast channel for broadcasting messages
	broadcast chan []byte
	// mu protects connections map
	mu sync.RWMutex
	// done signals hub shutdown
	done chan struct{}
}

// DefaultConfig returns the default WebSocket configuration
func DefaultConfig() Config {
	return Config{
		ReadBufferSize:  DefaultReadBufferSize,
		WriteBufferSize: DefaultWriteBufferSize,
		PingInterval:    DefaultPingInterval,
		PongTimeout:     DefaultPongTimeout,
		MaxMessageSize:  DefaultMaxMessageSize,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins by default
		},
	}
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan []byte, 256),
		done:        make(chan struct{}),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.connections[conn.ID] = conn
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[conn.ID]; ok {
				delete(h.connections, conn.ID)
				close(conn.sendChan)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, conn := range h.connections {
				select {
				case conn.sendChan <- message:
				default:
					// Buffer full, skip this connection
				}
			}
			h.mu.RUnlock()

		case <-h.done:
			return
		}
	}
}

// Stop stops the hub
func (h *Hub) Stop() {
	close(h.done)
}

// Register adds a connection to the hub
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister removes a connection from the hub
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %w", err)
	}
	h.broadcast <- data
	return nil
}

// BroadcastRaw sends raw bytes to all connected clients
func (h *Hub) BroadcastRaw(data []byte) {
	h.broadcast <- data
}

// Count returns the number of connected clients
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// GetConnection returns a connection by ID
func (h *Hub) GetConnection(id string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, ok := h.connections[id]
	return conn, ok
}

// ForEach iterates over all connections
func (h *Hub) ForEach(fn func(*Connection)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, conn := range h.connections {
		fn(conn)
	}
}

// Connection methods

// Send sends a message to the client
func (c *Connection) Send(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return c.SendRaw(data)
}

// SendJSON sends a JSON-encoded message
func (c *Connection) SendJSON(v interface{}) error {
	return c.Send(v)
}

// SendRaw sends raw bytes to the client
func (c *Connection) SendRaw(data []byte) error {
	c.closeMu.Lock()
	if c.closed {
		c.closeMu.Unlock()
		return errors.New("connection is closed")
	}
	c.closeMu.Unlock()

	select {
	case c.sendChan <- data:
		return nil
	default:
		return errors.New("send buffer full")
	}
}

// SendTyped sends a typed message with metadata
func (c *Connection) SendTyped(msgType string, payload interface{}) error {
	msg := Message[interface{}]{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	return c.Send(msg)
}

// ReadJSON reads and decodes a JSON message
func (c *Connection) ReadJSON(v interface{}) error {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Close closes the WebSocket connection
func (c *Connection) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	close(c.done)

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Context returns the Echo context
func (c *Connection) Context() echo.Context {
	return c.context
}

// SetMetadata sets a metadata value
func (c *Connection) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.metadata == nil {
		c.metadata = make(map[string]interface{})
	}
	c.metadata[key] = value
}

// GetMetadata gets a metadata value
func (c *Connection) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.metadata == nil {
		return nil, false
	}
	v, ok := c.metadata[key]
	return v, ok
}

// RemoteAddr returns the remote address
func (c *Connection) RemoteAddr() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return ""
}

// IsClosed returns whether the connection is closed
func (c *Connection) IsClosed() bool {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	return c.closed
}

// GenerateConnectionID generates a unique connection ID
func GenerateConnectionID() string {
	return fmt.Sprintf("ws-%d", time.Now().UnixNano())
}

// NewConnection creates a new WebSocket connection wrapper
func NewConnection(id string, ws *websocket.Conn, ctx echo.Context) *Connection {
	return &Connection{
		ID:       id,
		conn:     ws,
		context:  ctx,
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}
}

// CreateUpgrader creates a websocket.Upgrader from Config
func CreateUpgrader(config *Config) websocket.Upgrader {
	if config == nil {
		defaultConfig := DefaultConfig()
		config = &defaultConfig
	}
	return websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
	}
}

// HandleHandler handles Handler interface implementations
func HandleHandler(conn *Connection, handler Handler, config *Config) error {
	// Call OnConnect
	if err := handler.OnConnect(conn); err != nil {
		conn.Close()
		return err
	}

	// Start writer goroutine
	go Writer(conn, config)

	// Read loop
	var closeErr error
	for {
		messageType, data, err := conn.conn.ReadMessage()
		if err != nil {
			closeErr = err
			break
		}

		if err := handler.OnMessage(conn, messageType, data); err != nil {
			closeErr = err
			break
		}
	}

	// Call OnDisconnect
	handler.OnDisconnect(conn, closeErr)
	conn.Close()

	return nil
}

// HandleTypedHandler handles function-based handlers with typed messages
func HandleTypedHandler(conn *Connection, handler interface{}, config *Config) error {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	if handlerType.Kind() != reflect.Func {
		return errors.New("handler must be a function")
	}

	// Start writer goroutine
	go Writer(conn, config)

	// Determine message type from function signature
	var msgType reflect.Type
	if handlerType.NumIn() >= 2 {
		msgType = handlerType.In(1)
	}

	// Read loop
	for {
		_, data, err := conn.conn.ReadMessage()
		if err != nil {
			break
		}

		// Build arguments
		args := []reflect.Value{reflect.ValueOf(conn)}

		if msgType != nil {
			// Create new instance of message type
			msgPtr := reflect.New(msgType)
			if err := json.Unmarshal(data, msgPtr.Interface()); err != nil {
				continue // Skip invalid messages
			}
			args = append(args, msgPtr.Elem())
		}

		// Call handler
		results := handlerValue.Call(args)

		// Handle response if any
		if len(results) > 0 {
			// Check for error
			lastResult := results[len(results)-1]
			if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if !lastResult.IsNil() {
					break
				}
			}

			// Send response if first result is not error
			if len(results) >= 1 && !results[0].IsZero() {
				if results[0].Type() != reflect.TypeOf((*error)(nil)).Elem() {
					conn.Send(results[0].Interface())
				}
			}
		}
	}

	conn.Close()
	return nil
}

// Writer handles writing messages to the WebSocket
func Writer(conn *Connection, config *Config) {
	ticker := time.NewTicker(config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-conn.sendChan:
			if !ok {
				if conn.conn != nil {
					conn.conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			if conn.conn != nil {
				conn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			}

		case <-ticker.C:
			if conn.conn != nil {
				conn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}

		case <-conn.done:
			return
		}
	}
}

// SetupConnection configures a connection with read limits and pong handler
func SetupConnection(conn *Connection, config *Config) {
	if conn.conn == nil {
		return
	}
	conn.conn.SetReadLimit(config.MaxMessageSize)
	conn.conn.SetReadDeadline(time.Now().Add(config.PongTimeout))
	conn.conn.SetPongHandler(func(string) error {
		conn.conn.SetReadDeadline(time.Now().Add(config.PongTimeout))
		return nil
	})
}
