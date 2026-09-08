//go:build js

// Command stk-wasm is the browser client, compiled to WebAssembly. It runs the
// exact same Noise XX handshake and encrypted session as the native stk-client
// (internal/secure), over a browser WebSocket (internal/transport JSConn), and
// exposes a small JS API that the page wires to xterm.js.
//
// Build: GOOS=js GOARCH=wasm go build -o web/stk.wasm ./cmd/stk-wasm
package main

import (
	"syscall/js"

	"github.com/trustsentinel/stk/internal/secure"
	"github.com/trustsentinel/stk/internal/transport"
)

func main() {
	js.Global().Set("stkConnect", js.FuncOf(stkConnect))
	select {} // keep the wasm runtime alive for callbacks
}

// stkConnect(config) -> { send(str), close() }
// config: { hubURL, room, agentPub, onData(Uint8Array), onStatus(str), onClose() }
func stkConnect(this js.Value, args []js.Value) any {
	cfg := args[0]
	hubURL := cfg.Get("hubURL").String()
	room := cfg.Get("room").String()
	agentPub := cfg.Get("agentPub").String()
	onData := cfg.Get("onData")
	onStatus := cfg.Get("onStatus")
	onClose := cfg.Get("onClose")

	status := func(s string) {
		if onStatus.Truthy() {
			onStatus.Invoke(s)
		}
	}
	closed := func() {
		if onClose.Truthy() {
			onClose.Invoke()
		}
	}

	sendCh := make(chan []byte, 32)
	var sess *secure.Session

	go func() {
		if agentPub == "" {
			status("error: agent key required")
			closed()
			return
		}
		pin, derr := secure.DecodePublic(agentPub)
		if derr != nil {
			status("error: bad agent key")
			closed()
			return
		}

		status("connecting")
		enc := js.Global().Call("encodeURIComponent", room).String()
		conn, err := transport.DialWS(hubURL + "?role=client&room=" + enc)
		if err != nil {
			status("error: " + err.Error())
			closed()
			return
		}

		kp, err := secure.GenerateKeypair() // ephemeral client identity for the demo
		if err != nil {
			status("keygen error: " + err.Error())
			conn.Close()
			closed()
			return
		}

		status("handshaking")
		sess, err = secure.Handshake(conn, secure.Config{Static: kp, Initiator: true, PeerStatic: pin})
		if err != nil {
			status("auth failed: " + err.Error())
			conn.Close()
			closed()
			return
		}
		status("connected: " + secure.EncodePublic(sess.PeerStatic))

		// session -> terminal
		go func() {
			for {
				data, rerr := sess.Read()
				if len(data) > 0 && onData.Truthy() {
					onData.Invoke(bytesToUint8(data))
				}
				if rerr != nil {
					break
				}
			}
			status("closed")
			closed()
		}()

		// keystrokes -> session
		for b := range sendCh {
			if werr := sess.Write(b); werr != nil {
				break
			}
		}
	}()

	obj := js.Global().Get("Object").New()
	obj.Set("send", js.FuncOf(func(this js.Value, a []js.Value) any {
		if len(a) > 0 {
			select {
			case sendCh <- []byte(a[0].String()):
			default:
			}
		}
		return nil
	}))
	obj.Set("close", js.FuncOf(func(this js.Value, a []js.Value) any {
		if sess != nil {
			sess.Close()
		}
		return nil
	}))
	return obj
}

func bytesToUint8(b []byte) js.Value {
	u8 := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(u8, b)
	return u8
}
