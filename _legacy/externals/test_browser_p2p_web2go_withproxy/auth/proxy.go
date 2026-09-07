package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	flag.Parse()

	p := &Proxy{}
	log.Fatalf("%s", p.ListenAndServe("127.0.0.1:8081"))
}

// Proxy routes connections to backends based on a Config.
type Proxy struct {
	l net.Listener
}

// Serve accepts connections from l and routes them according to TLS SNI.
func (p *Proxy) Serve(l net.Listener) error {
	for {
		c, err := l.Accept()
		if err != nil {
			return fmt.Errorf("accept new conn: %s", err)
		}

		conn := &Conn{
			TCPConn: c.(*net.TCPConn),
		}
		go conn.proxy()
	}
}

// ListenAndServe creates a listener on addr calls Serve on it.
func (p *Proxy) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("create listener: %s", err)
	}
	return p.Serve(l)
}

// A Conn handles the TLS proxying of one user connection.
type Conn struct {
	*net.TCPConn

	tlsMinor    int
	hostname    string
	backend     string
	backendConn *net.TCPConn
}

func (c *Conn) logf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	log.Printf("%s <> %s: %s", c.RemoteAddr(), c.LocalAddr(), msg)
}

func (c *Conn) abort(alert byte, msg string, args ...interface{}) {
	c.logf(msg, args...)
	alertMsg := []byte{21, 3, byte(c.tlsMinor), 0, 2, 2, alert}

	if _, err := c.Write(alertMsg); err != nil {
		c.logf("error while sending alert: %s", err)
	}
}

func (c *Conn) internalError(msg string, args ...interface{}) { c.abort(80, msg, args...) }
func (c *Conn) sniFailed(msg string, args ...interface{})     { c.abort(112, msg, args...) }

func (c *Conn) proxy() {
	defer c.Close()

	var (
		err error
	)

	c.backend = "127.0.0.1:8080"
	c.logf("routing %q to %q", c.hostname, c.backend)
	backend, err := net.DialTimeout("tcp", c.backend, 10*time.Second)
	if err != nil {
		c.internalError("failed to dial backend %q for %q: %s", c.backend, c.hostname, err)
		return
	}
	defer backend.Close()

	c.backendConn = backend.(*net.TCPConn)

	var wg sync.WaitGroup
	wg.Add(2)
	go proxy(&wg, c.TCPConn, c.backendConn)
	go proxy(&wg, c.backendConn, c.TCPConn)
	wg.Wait()
}

func proxy(wg *sync.WaitGroup, a, b net.Conn) {
	defer wg.Done()
	atcp, btcp := a.(*net.TCPConn), b.(*net.TCPConn)
	if _, err := io.Copy(atcp, btcp); err != nil {
		log.Printf("%s<>%s -> %s<>%s: %s", atcp.RemoteAddr(), atcp.LocalAddr(), btcp.LocalAddr(), btcp.RemoteAddr(), err)
	}
	btcp.CloseWrite()
	atcp.CloseRead()
}
