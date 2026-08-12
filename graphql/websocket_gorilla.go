package graphql

import (
	"io"
	"net/http"
	"slices"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

// gorillaWebsocketImplementation adapts github.com/gorilla/websocket to the gqlgen
// websocket transport.
//
// gqlgen v0.17.94 replaced the transport's built-in gorilla Upgrader field with a
// pluggable Implementation backed by github.com/coder/websocket. This adapter keeps
// Config.WebSocketUpgrader working as before and keeps the framework on a single
// websocket library (see package websocket, which uses gorilla for the Hub).
//
// The gqlgen transport serialises every write through its own mutex, so the single
// concurrent writer that gorilla requires is preserved.
type gorillaWebsocketImplementation struct {
	upgrader websocket.Upgrader
}

var _ transport.WebsocketImplementation = gorillaWebsocketImplementation{}

// Accept upgrades the request, negotiating gqlgen's GraphQL subprotocols alongside
// any the caller configured on the upgrader.
func (g gorillaWebsocketImplementation) Accept(
	w http.ResponseWriter,
	r *http.Request,
	options transport.WebsocketAcceptOptions,
) (transport.WebsocketConn, error) {
	upgrader := g.upgrader

	// Clone before appending so a shared backing array is never mutated.
	subprotocols := slices.Clone(upgrader.Subprotocols)
	for _, subprotocol := range options.Subprotocols {
		if !slices.Contains(subprotocols, subprotocol) {
			subprotocols = append(subprotocols, subprotocol)
		}
	}
	upgrader.Subprotocols = subprotocols

	conn, err := upgrader.Upgrade(w, r, options.ResponseHeader)
	if err != nil {
		return nil, err
	}

	return &gorillaWebsocketConn{Conn: conn}, nil
}

// gorillaWebsocketConn satisfies transport.WebsocketConn. Close, WriteJSON,
// Subprotocol, SetReadLimit and SetReadDeadline are promoted from the embedded
// gorilla connection unchanged; only the two methods below need adapting.
type gorillaWebsocketConn struct {
	*websocket.Conn
}

var (
	_ transport.WebsocketConn          = (*gorillaWebsocketConn)(nil)
	_ transport.WebsocketReadLimiter   = (*gorillaWebsocketConn)(nil)
	_ transport.WebsocketReadDeadliner = (*gorillaWebsocketConn)(nil)
)

// NextReader reports a normal peer close as transport.ErrWebsocketClosed, the
// sentinel gqlgen uses to end a connection quietly rather than as an error.
func (c *gorillaWebsocketConn) NextReader() (int, io.Reader, error) {
	messageType, r, err := c.Conn.NextReader()
	if err != nil && websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseNoStatusReceived,
	) {
		return messageType, nil, transport.ErrWebsocketClosed
	}
	return messageType, r, err
}

// WriteClose sends a close frame. gorilla exposes no single-call equivalent, so
// the frame is formatted and written directly.
func (c *gorillaWebsocketConn) WriteClose(closeCode int, message string) error {
	return c.Conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(closeCode, message),
	)
}
