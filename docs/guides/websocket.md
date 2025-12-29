# WebSocket Support

EchoNext provides type-safe WebSocket handlers with connection management and broadcasting capabilities.

## Basic Usage

```go
import (
    "github.com/abdussamadbello/echonext"
    "github.com/abdussamadbello/echonext/websocket"
)

func chatHandler(conn *websocket.Connection) error {
    for {
        // Read message
        var msg ChatMessage
        if err := conn.ReadJSON(&msg); err != nil {
            return err // Connection closed
        }

        // Process and respond
        response := ChatResponse{
            Text:      "Echo: " + msg.Text,
            Timestamp: time.Now(),
        }

        if err := conn.WriteJSON(response); err != nil {
            return err
        }
    }
}

func main() {
    app := echonext.New()

    app.WS("/chat", chatHandler)

    app.Start(":8080")
}
```

## Connection Type

The `websocket.Connection` type provides:

```go
type Connection struct {
    ID       string                 // Unique connection ID
    Metadata map[string]interface{} // Custom metadata storage
}

// Methods
func (c *Connection) ReadJSON(v interface{}) error     // Read JSON message
func (c *Connection) WriteJSON(v interface{}) error    // Write JSON message
func (c *Connection) ReadMessage() (int, []byte, error) // Read raw message
func (c *Connection) WriteMessage(t int, data []byte) error // Write raw message
func (c *Connection) Close() error                     // Close connection

// Metadata
func (c *Connection) SetMetadata(key string, value interface{})
func (c *Connection) GetMetadata(key string) interface{}
```

## Hub Pattern

Use the Hub pattern for managing multiple connections and broadcasting:

```go
import "github.com/abdussamadbello/echonext/websocket"

type ChatHandler struct {
    hub *websocket.Hub
}

func NewChatHandler() *ChatHandler {
    hub := websocket.NewHub()
    go hub.Run() // Start hub in background
    return &ChatHandler{hub: hub}
}

func (h *ChatHandler) OnConnect(conn *websocket.Connection) error {
    // Store user info
    conn.SetMetadata("username", "anonymous")

    // Register with hub
    h.hub.Register(conn)

    // Broadcast join message
    h.hub.Broadcast(Message{
        Type: "join",
        Text: "A user joined the chat",
    })

    return nil
}

func (h *ChatHandler) OnMessage(conn *websocket.Connection, msgType int, data []byte) error {
    var msg ChatMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return err
    }

    // Broadcast to all connections
    response := Message{
        Type:     "message",
        Username: conn.GetMetadata("username").(string),
        Text:     msg.Text,
        Time:     time.Now(),
    }

    return h.hub.Broadcast(response)
}

func (h *ChatHandler) OnDisconnect(conn *websocket.Connection, err error) {
    h.hub.Unregister(conn)

    h.hub.Broadcast(Message{
        Type: "leave",
        Text: "A user left the chat",
    })
}

// Usage
handler := NewChatHandler()
app.WS("/ws/chat", handler)
```

## Hub Methods

```go
type Hub struct {
    // ...
}

func NewHub() *Hub                              // Create new hub
func (h *Hub) Run()                            // Start hub (run in goroutine)
func (h *Hub) Register(conn *Connection)        // Add connection
func (h *Hub) Unregister(conn *Connection)      // Remove connection
func (h *Hub) Broadcast(msg interface{}) error  // Send to all connections
func (h *Hub) BroadcastRaw(data []byte) error   // Send raw bytes to all
func (h *Hub) Count() int                       // Get connection count
func (h *Hub) GetConnection(id string) *Connection // Get specific connection
func (h *Hub) ForEach(fn func(*Connection))     // Iterate all connections
```

## Handler Interface

Implement the full handler interface for complete control:

```go
type WSHandler interface {
    OnConnect(conn *Connection) error
    OnMessage(conn *Connection, messageType int, data []byte) error
    OnDisconnect(conn *Connection, err error)
}
```

Or use a simple function:

```go
// Simple function handler
app.WS("/echo", func(conn *websocket.Connection) error {
    for {
        msgType, data, err := conn.ReadMessage()
        if err != nil {
            return err
        }
        if err := conn.WriteMessage(msgType, data); err != nil {
            return err
        }
    }
})
```

## Message Types

Define structured message types:

```go
type Message struct {
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}

type ChatMessage struct {
    Text string `json:"text"`
}

type UserJoinedPayload struct {
    Username string `json:"username"`
    UserID   string `json:"user_id"`
}
```

## Authentication

Authenticate WebSocket connections:

```go
func (h *ChatHandler) OnConnect(conn *websocket.Connection) error {
    // Get echo context from upgrade request
    // Authentication should happen before upgrade

    // Store authenticated user
    conn.SetMetadata("user_id", "user123")
    conn.SetMetadata("username", "john")

    h.hub.Register(conn)
    return nil
}
```

## Room/Channel Support

Implement rooms using multiple hubs or connection metadata:

```go
type RoomManager struct {
    rooms map[string]*websocket.Hub
    mu    sync.RWMutex
}

func (rm *RoomManager) GetOrCreateRoom(roomID string) *websocket.Hub {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    if hub, exists := rm.rooms[roomID]; exists {
        return hub
    }

    hub := websocket.NewHub()
    go hub.Run()
    rm.rooms[roomID] = hub
    return hub
}

func (h *ChatHandler) OnConnect(conn *websocket.Connection) error {
    roomID := conn.GetMetadata("room_id").(string)
    room := h.roomManager.GetOrCreateRoom(roomID)
    room.Register(conn)
    return nil
}
```

## Error Handling

```go
func (h *ChatHandler) OnMessage(conn *websocket.Connection, msgType int, data []byte) error {
    var msg Message
    if err := json.Unmarshal(data, &msg); err != nil {
        // Send error back to client
        conn.WriteJSON(Message{
            Type:  "error",
            Error: "Invalid message format",
        })
        return nil // Don't disconnect on parse errors
    }

    switch msg.Type {
    case "chat":
        return h.handleChat(conn, msg)
    case "ping":
        return conn.WriteJSON(Message{Type: "pong"})
    default:
        conn.WriteJSON(Message{
            Type:  "error",
            Error: "Unknown message type",
        })
        return nil
    }
}
```

## Client-Side JavaScript

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/chat');

ws.onopen = () => {
    console.log('Connected');
    ws.send(JSON.stringify({
        type: 'chat',
        text: 'Hello!'
    }));
};

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received:', msg);
};

ws.onclose = () => {
    console.log('Disconnected');
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};
```

## CLI Generator

Generate WebSocket handler boilerplate:

```bash
echonext generate websocket chat
```

This creates:

```
internal/ws/chat/
├── handler.go     # WebSocket connection handler
├── hub.go         # Connection management and broadcasting
└── message.go     # Message types and serialization
```

## Best Practices

1. **Use the Hub pattern** - Centralize connection management
2. **Handle disconnects gracefully** - Clean up resources in OnDisconnect
3. **Implement heartbeats** - Keep connections alive with ping/pong
4. **Rate limit messages** - Prevent abuse from rapid message sending
5. **Validate all input** - Never trust client messages
6. **Use JSON for messages** - Structured data is easier to handle
7. **Log connection events** - Track connects/disconnects for debugging

## Heartbeat Example

```go
func (h *ChatHandler) OnConnect(conn *websocket.Connection) error {
    h.hub.Register(conn)

    // Start heartbeat
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            if err := conn.WriteJSON(Message{Type: "ping"}); err != nil {
                return
            }
        }
    }()

    return nil
}
```

## Example Project

See [examples/websocket-demo/](../../examples/websocket-demo/) for a complete working example with:

- Real-time chat application
- Hub pattern implementation
- HTML client interface
- Message broadcasting
- Connection management
