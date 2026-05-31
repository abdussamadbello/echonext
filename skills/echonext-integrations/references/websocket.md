# WebSocket — Reference

Package `github.com/abdussamadbello/echonext/websocket`.

## Handler interface

The framework calls these on the connection lifecycle (see
`websocket/websocket.go`). `OnMessage` receives the gorilla-style
`messageType` (e.g. text/binary) plus the raw payload:

```go
type Handler interface {
    OnConnect(conn *websocket.Connection) error
    OnMessage(conn *websocket.Connection, messageType int, data []byte) error
    OnDisconnect(conn *websocket.Connection, err error)
}
```

## Connection

```go
func (c *Connection) Send(message interface{}) error  // JSON-encodes
func (c *Connection) SendJSON(v interface{}) error
func (c *Connection) SendRaw(data []byte) error
func (c *Connection) Close() error
func (c *Connection) SetMetadata(key string, value interface{})
func (c *Connection) GetMetadata(key string) (interface{}, bool)
```

## Hub

```go
func NewHub() *Hub
func (h *Hub) Run()                       // run in a goroutine
func (h *Hub) Register(conn *Connection)
func (h *Hub) Unregister(conn *Connection)
func (h *Hub) Broadcast(message interface{}) error
func (h *Hub) Count() int
func (h *Hub) GetConnection(id string) (*Connection, bool)
func (h *Hub) ForEach(fn func(*Connection))
```

## Config

```go
type Config struct {
    ReadBufferSize  int
    WriteBufferSize int
    PingInterval    time.Duration
    PongTimeout     time.Duration
    MaxMessageSize  int64
    CheckOrigin     func(r *http.Request) bool
}
```

## Chat example

```go
package chat

import (
    "encoding/json"

    "github.com/abdussamadbello/echonext/websocket"
)

type Message struct {
    Type    string `json:"type"`
    Payload string `json:"payload"`
}

type Handler struct {
    hub *websocket.Hub
}

func NewHandler() *Handler {
    h := &Handler{hub: websocket.NewHub()}
    go h.hub.Run()
    return h
}

func (h *Handler) OnConnect(conn *websocket.Connection) error {
    h.hub.Register(conn)
    return conn.SendJSON(Message{Type: "system", Payload: "welcome"})
}

func (h *Handler) OnMessage(conn *websocket.Connection, messageType int, data []byte) error {
    var in Message
    if err := json.Unmarshal(data, &in); err != nil {
        return err
    }
    return h.hub.Broadcast(Message{Type: "message", Payload: in.Payload})
}

func (h *Handler) OnDisconnect(conn *websocket.Connection, err error) {
    h.hub.Unregister(conn)
}
```

Register it:

```go
app.WS("/ws/chat", chat.NewHandler(), echonext.Route{
    Summary: "Chat WebSocket",
    Tags:    []string{"Chat"},
})
```

Scaffold the boilerplate with `echonext generate websocket chat`.
