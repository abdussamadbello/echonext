package websocket

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test message types
type ChatMessage struct {
	Text string `json:"text"`
	From string `json:"from"`
}

type ChatResponse struct {
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

func TestHub_Basic(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create mock connections
	conn1 := &Connection{
		ID:       "conn-1",
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}
	conn2 := &Connection{
		ID:       "conn-2",
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	// Register connections
	hub.Register(conn1)
	hub.Register(conn2)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.Count())

	// Unregister one
	hub.Unregister(conn1)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 1, hub.Count())
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create connections with buffered channels
	conn1 := &Connection{
		ID:       "conn-1",
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}
	conn2 := &Connection{
		ID:       "conn-2",
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	hub.Register(conn1)
	hub.Register(conn2)
	time.Sleep(50 * time.Millisecond)

	// Broadcast message
	msg := map[string]string{"type": "test", "message": "hello"}
	err := hub.Broadcast(msg)
	assert.NoError(t, err)

	// Both connections should receive the message
	time.Sleep(50 * time.Millisecond)

	select {
	case data := <-conn1.sendChan:
		assert.Contains(t, string(data), "hello")
	case <-time.After(100 * time.Millisecond):
		t.Error("conn1 didn't receive broadcast")
	}

	select {
	case data := <-conn2.sendChan:
		assert.Contains(t, string(data), "hello")
	case <-time.After(100 * time.Millisecond):
		t.Error("conn2 didn't receive broadcast")
	}
}

func TestHub_GetConnection(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	conn := &Connection{
		ID:       "test-123",
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	hub.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Get existing connection
	found, ok := hub.GetConnection("test-123")
	assert.True(t, ok)
	assert.Equal(t, conn.ID, found.ID)

	// Get non-existent connection
	_, ok = hub.GetConnection("non-existent")
	assert.False(t, ok)
}

func TestHub_ForEach(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	conn1 := &Connection{ID: "conn-1", metadata: make(map[string]interface{}), sendChan: make(chan []byte, 256), done: make(chan struct{})}
	conn2 := &Connection{ID: "conn-2", metadata: make(map[string]interface{}), sendChan: make(chan []byte, 256), done: make(chan struct{})}
	conn3 := &Connection{ID: "conn-3", metadata: make(map[string]interface{}), sendChan: make(chan []byte, 256), done: make(chan struct{})}

	hub.Register(conn1)
	hub.Register(conn2)
	hub.Register(conn3)
	time.Sleep(50 * time.Millisecond)

	var ids []string
	var mu sync.Mutex
	hub.ForEach(func(conn *Connection) {
		mu.Lock()
		ids = append(ids, conn.ID)
		mu.Unlock()
	})

	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "conn-1")
	assert.Contains(t, ids, "conn-2")
	assert.Contains(t, ids, "conn-3")
}

func TestConnection_Metadata(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		metadata: make(map[string]interface{}),
	}

	// Set metadata
	conn.SetMetadata("user_id", 123)
	conn.SetMetadata("role", "admin")
	conn.SetMetadata("tags", []string{"vip", "premium"})

	// Get metadata
	userID, ok := conn.GetMetadata("user_id")
	assert.True(t, ok)
	assert.Equal(t, 123, userID)

	role, ok := conn.GetMetadata("role")
	assert.True(t, ok)
	assert.Equal(t, "admin", role)

	tags, ok := conn.GetMetadata("tags")
	assert.True(t, ok)
	assert.Equal(t, []string{"vip", "premium"}, tags)

	// Non-existent key
	_, ok = conn.GetMetadata("non-existent")
	assert.False(t, ok)
}

func TestConnection_Send(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	// Send message
	msg := ChatMessage{Text: "Hello", From: "user"}
	err := conn.Send(msg)
	assert.NoError(t, err)

	// Verify message was queued
	select {
	case data := <-conn.sendChan:
		assert.Contains(t, string(data), "Hello")
		assert.Contains(t, string(data), "user")
	case <-time.After(100 * time.Millisecond):
		t.Error("Message was not queued")
	}
}

func TestConnection_SendTyped(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	// Send typed message
	payload := ChatMessage{Text: "Hello", From: "server"}
	err := conn.SendTyped("chat", payload)
	assert.NoError(t, err)

	// Verify message
	select {
	case data := <-conn.sendChan:
		assert.Contains(t, string(data), "chat")
		assert.Contains(t, string(data), "Hello")
		assert.Contains(t, string(data), "timestamp")
	case <-time.After(100 * time.Millisecond):
		t.Error("Message was not queued")
	}
}

func TestConnection_Close(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		done:     make(chan struct{}),
		sendChan: make(chan []byte, 256),
	}

	// First close should succeed
	err := conn.Close()
	assert.NoError(t, err)
	assert.True(t, conn.closed)

	// Second close should also succeed (no-op)
	err = conn.Close()
	assert.NoError(t, err)
}

func TestConnection_SendAfterClose(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		done:     make(chan struct{}),
		sendChan: make(chan []byte, 256),
	}

	conn.Close()

	err := conn.SendRaw([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestConnection_SendBufferFull(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		done:     make(chan struct{}),
		sendChan: make(chan []byte, 1), // Small buffer
	}

	// Fill the buffer
	conn.sendChan <- []byte("first")

	// Next send should fail
	err := conn.SendRaw([]byte("second"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "buffer full")
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, DefaultReadBufferSize, config.ReadBufferSize)
	assert.Equal(t, DefaultWriteBufferSize, config.WriteBufferSize)
	assert.Equal(t, DefaultPingInterval, config.PingInterval)
	assert.Equal(t, DefaultPongTimeout, config.PongTimeout)
	assert.Equal(t, int64(DefaultMaxMessageSize), config.MaxMessageSize)
	assert.NotNil(t, config.CheckOrigin)

	// Default CheckOrigin should allow all
	req := &http.Request{}
	assert.True(t, config.CheckOrigin(req))
}

func TestMessage(t *testing.T) {
	msg := Message[ChatMessage]{
		Type: "chat",
		Payload: ChatMessage{
			Text: "Hello",
			From: "user",
		},
		Timestamp: time.Now(),
		ID:        "msg-123",
	}

	assert.Equal(t, "chat", msg.Type)
	assert.Equal(t, "Hello", msg.Payload.Text)
	assert.Equal(t, "user", msg.Payload.From)
	assert.NotZero(t, msg.Timestamp)
	assert.Equal(t, "msg-123", msg.ID)
}

func TestRoute(t *testing.T) {
	route := Route{
		Summary:     "Chat WebSocket",
		Description: "WebSocket endpoint for chat",
		Tags:        []string{"chat", "websocket"},
		Protocols:   []string{"chat.v1", "chat.v2"},
	}

	assert.Equal(t, "Chat WebSocket", route.Summary)
	assert.Equal(t, "WebSocket endpoint for chat", route.Description)
	assert.Len(t, route.Tags, 2)
	assert.Len(t, route.Protocols, 2)
}

func TestGenerateConnectionID(t *testing.T) {
	id1 := GenerateConnectionID()
	id2 := GenerateConnectionID()

	// IDs should be unique
	assert.NotEqual(t, id1, id2)

	// IDs should have the expected prefix
	assert.Contains(t, id1, "ws-")
	assert.Contains(t, id2, "ws-")
}

func TestHub_BroadcastRaw(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	conn := &Connection{
		ID:       "test",
		metadata: make(map[string]interface{}),
		sendChan: make(chan []byte, 256),
		done:     make(chan struct{}),
	}

	hub.Register(conn)
	time.Sleep(50 * time.Millisecond)

	// Broadcast raw bytes
	hub.BroadcastRaw([]byte("raw message"))

	select {
	case data := <-conn.sendChan:
		assert.Equal(t, "raw message", string(data))
	case <-time.After(100 * time.Millisecond):
		t.Error("Didn't receive raw broadcast")
	}
}

func TestConnection_IsClosed(t *testing.T) {
	conn := &Connection{
		ID:       "test",
		done:     make(chan struct{}),
		sendChan: make(chan []byte, 256),
	}

	assert.False(t, conn.IsClosed())

	conn.Close()

	assert.True(t, conn.IsClosed())
}

func TestNewConnection(t *testing.T) {
	conn := NewConnection("test-id", nil, nil)

	assert.Equal(t, "test-id", conn.ID)
	assert.NotNil(t, conn.metadata)
	assert.NotNil(t, conn.sendChan)
	assert.NotNil(t, conn.done)
}

func TestCreateUpgrader(t *testing.T) {
	// Test with nil config
	upgrader := CreateUpgrader(nil)
	assert.Equal(t, DefaultReadBufferSize, upgrader.ReadBufferSize)
	assert.Equal(t, DefaultWriteBufferSize, upgrader.WriteBufferSize)

	// Test with custom config
	config := &Config{
		ReadBufferSize:  2048,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return false
		},
	}
	upgrader = CreateUpgrader(config)
	assert.Equal(t, 2048, upgrader.ReadBufferSize)
	assert.Equal(t, 4096, upgrader.WriteBufferSize)
}

// Benchmark tests
func BenchmarkHub_Register(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := &Connection{
			ID:       GenerateConnectionID(),
			metadata: make(map[string]interface{}),
			sendChan: make(chan []byte, 256),
			done:     make(chan struct{}),
		}
		hub.Register(conn)
	}
}

func BenchmarkHub_Broadcast(b *testing.B) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Register some connections
	for i := 0; i < 100; i++ {
		conn := &Connection{
			ID:       GenerateConnectionID(),
			metadata: make(map[string]interface{}),
			sendChan: make(chan []byte, 256),
			done:     make(chan struct{}),
		}
		hub.Register(conn)
	}
	time.Sleep(100 * time.Millisecond)

	msg := map[string]string{"type": "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.Broadcast(msg)
	}
}

func BenchmarkConnection_Send(b *testing.B) {
	conn := &Connection{
		ID:       "test",
		sendChan: make(chan []byte, 10000),
		done:     make(chan struct{}),
	}

	msg := ChatMessage{Text: "benchmark", From: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.Send(msg)
	}
}
