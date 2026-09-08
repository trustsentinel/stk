//go:build js

package transport

import (
	"errors"
	"sync"
	"syscall/js"
)

// JSConn implements MsgConn over a browser WebSocket (binary/arraybuffer frames),
// so the same secure.Session code runs in the browser under GOOS=js/wasm.
type JSConn struct {
	ws        js.Value
	in        chan []byte
	closed    chan struct{}
	once      sync.Once
	onMessage js.Func
	onClose   js.Func
	onError   js.Func
}

// DialWS opens a WebSocket to url and returns once it is open.
func DialWS(url string) (*JSConn, error) {
	ws := js.Global().Get("WebSocket").New(url)
	ws.Set("binaryType", "arraybuffer")
	c := &JSConn{
		ws:     ws,
		in:     make(chan []byte, 32),
		closed: make(chan struct{}),
	}

	openCh := make(chan error, 1)
	onOpen := js.FuncOf(func(this js.Value, args []js.Value) any {
		select {
		case openCh <- nil:
		default:
		}
		return nil
	})
	c.onMessage = js.FuncOf(func(this js.Value, args []js.Value) any {
		buf := arrayBufferToBytes(args[0].Get("data"))
		select {
		case c.in <- buf:
		case <-c.closed:
		}
		return nil
	})
	c.onError = js.FuncOf(func(this js.Value, args []js.Value) any {
		select {
		case openCh <- errors.New("transport: websocket error"):
		default:
		}
		c.Close()
		return nil
	})
	c.onClose = js.FuncOf(func(this js.Value, args []js.Value) any {
		c.Close()
		return nil
	})

	ws.Call("addEventListener", "open", onOpen)
	ws.Call("addEventListener", "message", c.onMessage)
	ws.Call("addEventListener", "error", c.onError)
	ws.Call("addEventListener", "close", c.onClose)

	err := <-openCh // blocks this goroutine only; the JS event loop keeps running
	onOpen.Release()
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// WriteMsg sends p as one binary WebSocket frame.
func (c *JSConn) WriteMsg(p []byte) error {
	select {
	case <-c.closed:
		return errClosed
	default:
	}
	u8 := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(u8, p)
	c.ws.Call("send", u8)
	return nil
}

// ReadMsg returns the next binary frame's payload.
func (c *JSConn) ReadMsg() ([]byte, error) {
	select {
	case b := <-c.in:
		return b, nil
	case <-c.closed:
		return nil, errClosed
	}
}

// Close closes the WebSocket and releases callbacks.
func (c *JSConn) Close() error {
	c.once.Do(func() {
		close(c.closed)
		c.ws.Call("close")
		c.onMessage.Release()
		c.onClose.Release()
		c.onError.Release()
	})
	return nil
}

func arrayBufferToBytes(ab js.Value) []byte {
	u8 := js.Global().Get("Uint8Array").New(ab)
	n := u8.Get("length").Int()
	b := make([]byte, n)
	js.CopyBytesToGo(b, u8)
	return b
}
