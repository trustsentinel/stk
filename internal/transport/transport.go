// Package transport provides a small message-framed connection abstraction.
//
// stk relays whole messages (Noise handshake messages, then encrypted transport
// frames) rather than a raw byte stream, so the unit everything is built on is a
// MsgConn: send one []byte, receive one []byte. A WebSocket (one binary frame per
// message) and an in-memory Pipe (for tests) both implement it, which lets the
// secure session and the hub relay be tested without a network.
package transport

import (
	"sync"

	"github.com/gorilla/websocket"
)

// MsgConn is a bidirectional, message-oriented connection: each WriteMsg is
// delivered to the peer as exactly one ReadMsg.
type MsgConn interface {
	WriteMsg(p []byte) error
	ReadMsg() ([]byte, error)
	Close() error
}

// WSConn adapts a gorilla *websocket.Conn to MsgConn using binary frames.
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

// pipeConn is one end of an in-memory MsgConn pair.
type pipeConn struct {
	in     chan []byte
	out    chan []byte
	closed chan struct{}
	once   sync.Once
}

// Pipe returns two connected in-memory MsgConns. Writes on one are readable on
// the other. Used by tests to exercise handshakes and relays without a network.
func Pipe() (MsgConn, MsgConn) {
	a2b := make(chan []byte, 16)
	b2a := make(chan []byte, 16)
	closed := make(chan struct{})
	a := &pipeConn{in: b2a, out: a2b, closed: closed}
	b := &pipeConn{in: a2b, out: b2a, closed: closed}
	return a, b
}

func (p *pipeConn) WriteMsg(b []byte) error {
	cp := make([]byte, len(b))
	copy(cp, b)
	select {
	case p.out <- cp:
		return nil
	case <-p.closed:
		return errClosed
	}
}

func (p *pipeConn) ReadMsg() ([]byte, error) {
	select {
	case b := <-p.in:
		return b, nil
	case <-p.closed:
		return nil, errClosed
	}
}

func (p *pipeConn) Close() error {
	p.once.Do(func() { close(p.closed) })
	return nil
}

type pipeError string

func (e pipeError) Error() string { return string(e) }

const errClosed = pipeError("transport: pipe closed")
