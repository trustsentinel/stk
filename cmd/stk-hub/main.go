// Command stk-hub is the broker. It pairs a client and an agent by room and
// relays opaque messages between them. Because the client and agent run an
// end-to-end Noise session on top, the hub only ever sees ciphertext — it holds
// no keys and cannot read the session. It is the only component that listens for
// inbound connections; the agent dials out to it.
package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// The browser client connects cross-origin; the session is authenticated and
	// encrypted end-to-end, so origin is not the security boundary here.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type broker struct {
	mu      sync.Mutex
	waiting map[string]*websocket.Conn // room -> agent awaiting a client
}

func (b *broker) handle(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	switch role {
	case "agent":
		b.mu.Lock()
		if old := b.waiting[room]; old != nil {
			old.Close()
		}
		b.waiting[room] = c
		b.mu.Unlock()
		log.Printf("agent registered (room=%s), waiting for a client", room)
		// The relay begins when a client arrives; until then this conn sits idle.
	case "client":
		b.mu.Lock()
		a := b.waiting[room]
		delete(b.waiting, room)
		b.mu.Unlock()
		if a == nil {
			log.Printf("client arrived but no agent for room=%s", room)
			c.Close()
			return
		}
		log.Printf("pairing client<->agent (room=%s); relaying opaque frames — hub sees only ciphertext", room)
		relay(a, c)
		log.Printf("session closed (room=%s)", room)
	default:
		c.Close()
	}
}

// relay copies messages both directions until either side closes.
func relay(a, c *websocket.Conn) {
	done := make(chan struct{}, 2)
	go copyMsgs(a, c, done) // client -> agent
	go copyMsgs(c, a, done) // agent -> client
	<-done
	a.Close()
	c.Close()
	<-done
}

func copyMsgs(dst, src *websocket.Conn, done chan struct{}) {
	for {
		t, msg, err := src.ReadMessage()
		if err != nil {
			break
		}
		if err := dst.WriteMessage(t, msg); err != nil {
			break
		}
	}
	done <- struct{}{}
}

func main() {
	addr := flag.String("addr", ":8443", "listen address")
	flag.Parse()

	b := &broker{waiting: map[string]*websocket.Conn{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", b.handle)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	log.Printf("stk-hub listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
