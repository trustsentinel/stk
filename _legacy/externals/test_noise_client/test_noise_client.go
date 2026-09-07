package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"websocket"

	"github.com/flynn/noise"
	"gopkg.in/noisesocket.v0"
)

func main() {
	CreateDedicatedChannel("localhost", 8100)
}

// CreateDedicatedChannel creates the secure channel using noise protocol framework
func CreateDedicatedChannel(address string, port int32) {
	pub1, _ := base64.StdEncoding.DecodeString("L9Xm5qy17ZZ6rBMd1Dsn5iZOyS7vUVhYK+zby1nJPEE=")
	priv1, _ := base64.StdEncoding.DecodeString("TPmwb3vTEgrA3oq6PoGEzH5hT91IDXGC9qEMc8ksRiw=")

	fmt.Println("client ss.pub", pub1)
	fmt.Println("client ss.priv", priv1)

	// Diffie Hellman pair keys
	clientKeys := noise.DHKey{
		Public:  pub1,
		Private: priv1,
	}

	u := url.URL{Scheme: "ws", Host: address + ":" + strconv.FormatInt(int64(port), 10), Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	// Websockets wrap a noise socket connection
	d := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		NetDial: func(network, addr string) (net.Conn, error) {
			conn, err := noisesocket.Dial(addr, &noisesocket.ConnectionConfig{StaticKey: clientKeys})
			if err != nil {
				fmt.Println("Dial", err)
			}
			return conn, err
		},
	}
	c, _, err := d.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()
	for {
		select {
		case <-done:
			return
		}
	}
}
