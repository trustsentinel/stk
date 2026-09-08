//go:build !js

package transport

import (
	"sync"

	"github.com/gorilla/websocket"
)

// WSConn adapts a gorilla *websocket.Conn to MsgConn using binary frames.
// Used by the hub, agent, and CLI client (native builds).
type WSConn struct {
	c  *websocket.Conn
	mu sync.Mutex // gorilla allows one concurrent writer; serialize writes
}

// NewWSConn wraps a websocket connection.
func NewWSConn(c *websocket.Conn) *WSConn { return &WSConn{c: c} }

// WriteMsg sends p as a single binary websocket frame.
func (w *WSConn) WriteMsg(p []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.c.WriteMessage(websocket.BinaryMessage, p)
}

// ReadMsg returns the payload of the next websocket data frame.
func (w *WSConn) ReadMsg() ([]byte, error) {
	_, p, err := w.c.ReadMessage()
	return p, err
}

// Close closes the underlying websocket.
func (w *WSConn) Close() error { return w.c.Close() }
