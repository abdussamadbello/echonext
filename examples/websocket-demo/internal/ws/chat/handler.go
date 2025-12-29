package chat

import (
	"encoding/json"
	"log"
	"time"

	"github.com/abdussamadbello/echonext/websocket"
)

// Handler implements the WebSocket handler interface for chat
type Handler struct {
	hub *Hub
}

// NewHandler creates a new Chat WebSocket handler
func NewHandler() *Handler {
	hub := NewHub()
	go hub.Run()
	return &Handler{hub: hub}
}

// OnConnect is called when a new client connects
func (h *Handler) OnConnect(conn *websocket.Connection) error {
	log.Printf("[chat] Client connected: %s", conn.ID)
	h.hub.Register(conn)

	// Send welcome message
	welcome := Message{
		Type:      MessageTypeSystem,
		Payload:   "Welcome to chat!",
		Timestamp: time.Now(),
	}
	return conn.SendJSON(welcome)
}

// OnMessage handles incoming messages from clients
func (h *Handler) OnMessage(conn *websocket.Connection, data []byte) error {
	var msg IncomingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return conn.SendJSON(Message{
			Type:      MessageTypeError,
			Payload:   "Invalid message format",
			Timestamp: time.Now(),
		})
	}

	log.Printf("[chat] Message from %s: %s", conn.ID, msg.Content)

	// Broadcast message to all connected clients
	response := Message{
		Type:      MessageTypeMessage,
		Payload:   msg.Content,
		Sender:    conn.ID,
		Timestamp: time.Now(),
	}

	return h.hub.Broadcast(response)
}

// OnDisconnect is called when a client disconnects
func (h *Handler) OnDisconnect(conn *websocket.Connection, err error) {
	log.Printf("[chat] Client disconnected: %s (error: %v)", conn.ID, err)
	h.hub.Unregister(conn)
}

// GetHub returns the connection hub (useful for external broadcasting)
func (h *Handler) GetHub() *Hub {
	return h.hub
}
