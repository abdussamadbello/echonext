package graphql

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// acceptOnce serves a single websocket upgrade through the adapter and hands the
// resulting connection to fn.
func acceptOnce(
	t *testing.T,
	upgrader websocket.Upgrader,
	options transport.WebsocketAcceptOptions,
	fn func(conn transport.WebsocketConn),
) *httptest.Server {
	t.Helper()

	impl := gorillaWebsocketImplementation{upgrader: upgrader}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := impl.Accept(w, r, options)
		if err != nil {
			return
		}
		defer conn.Close()
		fn(conn)
	}))
}

func dial(t *testing.T, server *httptest.Server, subprotocols []string) *websocket.Conn {
	t.Helper()

	dialer := websocket.Dialer{
		Subprotocols:     subprotocols,
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := dialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	require.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
	}
	return conn
}

func TestGorillaWebsocketImplementation_NegotiatesGqlgenSubprotocol(t *testing.T) {
	negotiated := make(chan string, 1)
	server := acceptOnce(t,
		websocket.Upgrader{},
		transport.WebsocketAcceptOptions{Subprotocols: []string{"graphql-transport-ws", "graphql-ws"}},
		func(conn transport.WebsocketConn) {
			negotiated <- conn.Subprotocol()
		},
	)
	defer server.Close()

	client := dial(t, server, []string{"graphql-transport-ws"})
	defer client.Close()

	assert.Equal(t, "graphql-transport-ws", <-negotiated,
		"gqlgen's subprotocols must be offered even when the upgrader configures none")
}

func TestGorillaWebsocketImplementation_KeepsConfiguredSubprotocols(t *testing.T) {
	negotiated := make(chan string, 1)
	server := acceptOnce(t,
		websocket.Upgrader{Subprotocols: []string{"custom-protocol"}},
		transport.WebsocketAcceptOptions{Subprotocols: []string{"graphql-transport-ws"}},
		func(conn transport.WebsocketConn) {
			negotiated <- conn.Subprotocol()
		},
	)
	defer server.Close()

	client := dial(t, server, []string{"custom-protocol"})
	defer client.Close()

	assert.Equal(t, "custom-protocol", <-negotiated,
		"subprotocols configured on the upgrader must survive the merge")
}

func TestGorillaWebsocketImplementation_DoesNotMutateCallerSubprotocols(t *testing.T) {
	// Extra capacity means an unguarded append would write into the caller's
	// backing array rather than a copy.
	configured := make([]string, 1, 8)
	configured[0] = "custom-protocol"
	shared := configured[:1:8]

	done := make(chan struct{})
	server := acceptOnce(t,
		websocket.Upgrader{Subprotocols: configured},
		transport.WebsocketAcceptOptions{Subprotocols: []string{"graphql-transport-ws", "graphql-ws"}},
		func(conn transport.WebsocketConn) { close(done) },
	)
	defer server.Close()

	client := dial(t, server, []string{"custom-protocol"})
	defer client.Close()
	<-done

	assert.Equal(t, []string{"custom-protocol"}, configured)

	// Had the merge appended in place, gqlgen's subprotocols would be sitting in
	// the caller's spare capacity instead of in a copy.
	backing := shared[:cap(shared)]
	assert.Equal(t, "", backing[1],
		"the adapter must not append into the caller's backing array")
	assert.Equal(t, "", backing[2],
		"the adapter must not append into the caller's backing array")
}

func TestGorillaWebsocketConn_NextReaderReportsNormalClose(t *testing.T) {
	readErr := make(chan error, 1)
	server := acceptOnce(t,
		websocket.Upgrader{},
		transport.WebsocketAcceptOptions{},
		func(conn transport.WebsocketConn) {
			_, _, err := conn.NextReader()
			readErr <- err
		},
	)
	defer server.Close()

	client := dial(t, server, nil)
	require.NoError(t, client.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	))
	defer client.Close()

	assert.ErrorIs(t, <-readErr, transport.ErrWebsocketClosed,
		"a normal peer close must surface as gqlgen's sentinel, not a raw error")
}

func TestGorillaWebsocketConn_NextReaderPassesThroughMessages(t *testing.T) {
	type read struct {
		messageType int
		payload     string
		err         error
	}
	got := make(chan read, 1)
	server := acceptOnce(t,
		websocket.Upgrader{},
		transport.WebsocketAcceptOptions{},
		func(conn transport.WebsocketConn) {
			messageType, r, err := conn.NextReader()
			if err != nil {
				got <- read{err: err}
				return
			}
			buf := make([]byte, 64)
			n, _ := r.Read(buf)
			got <- read{messageType: messageType, payload: string(buf[:n])}
		},
	)
	defer server.Close()

	client := dial(t, server, nil)
	defer client.Close()
	require.NoError(t, client.WriteMessage(websocket.TextMessage, []byte("hello")))

	result := <-got
	require.NoError(t, result.err)
	assert.Equal(t, websocket.TextMessage, result.messageType)
	assert.Equal(t, "hello", result.payload)
}

func TestGorillaWebsocketConn_WriteClose(t *testing.T) {
	server := acceptOnce(t,
		websocket.Upgrader{},
		transport.WebsocketAcceptOptions{},
		func(conn transport.WebsocketConn) {
			assert.NoError(t, conn.WriteClose(transport.WebsocketCloseProtocolError, "unsupported"))
		},
	)
	defer server.Close()

	client := dial(t, server, nil)
	defer client.Close()

	_, _, err := client.ReadMessage()
	assert.True(t,
		websocket.IsCloseError(err, transport.WebsocketCloseProtocolError),
		"expected close frame with the protocol-error code, got %v", err)
}

func TestGorillaWebsocketConn_WriteJSON(t *testing.T) {
	server := acceptOnce(t,
		websocket.Upgrader{},
		transport.WebsocketAcceptOptions{},
		func(conn transport.WebsocketConn) {
			assert.NoError(t, conn.WriteJSON(map[string]string{"type": "connection_ack"}))
		},
	)
	defer server.Close()

	client := dial(t, server, nil)
	defer client.Close()

	var payload map[string]string
	require.NoError(t, client.ReadJSON(&payload))
	assert.Equal(t, map[string]string{"type": "connection_ack"}, payload)
}

func TestGorillaWebsocketImplementation_AcceptSendsResponseHeader(t *testing.T) {
	server := acceptOnce(t,
		websocket.Upgrader{},
		transport.WebsocketAcceptOptions{
			ResponseHeader: http.Header{"X-Echonext": []string{"adapter"}},
		},
		func(conn transport.WebsocketConn) {},
	)
	defer server.Close()

	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, resp, err := dialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	require.NoError(t, err)
	defer conn.Close()
	defer resp.Body.Close()

	assert.Equal(t, "adapter", resp.Header.Get("X-Echonext"))
}
