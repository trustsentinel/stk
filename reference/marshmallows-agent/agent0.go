package main

import (
	"fmt"
	"os/exec"
	"websocket"

	"github.com/kr/pty"
)

func pttyAttach(c *websocket.Conn) {
	cmd := exec.Command("bash")
	f, _ := pty.Start(cmd)
	wc := &rwc{c: c}
	p := &Pipe{}
	go p.in(f, wc)
	p.out(wc, f)
}

func main2() {
	token := GenerateToken(6)
	fmt.Println("[agent]token", token)

	createClientPlainChannel("localhost", 9999, "/ws", func(in []byte, size int) []byte { return in },
		func(conn *websocket.Conn) {
			pttyAttach(conn)
		})

	/* complete flow
	// authorisation channel
	shutdown := make(chan bool)
	done := make(chan struct{})
	conn := authConnection("localhost", 8081, "/open/"+token) // @todo: ride off placeholder
	defer conn.Close()
	go func() {
		defer close(done)
		authHandler(token, conn, shutdown)
	}()

	// builds an interval signal used as Hearbeat
	healthcheck := time.NewTicker(10 * time.Second)
	defer healthcheck.Stop()

	for {
		select {
		case <-shutdown:
			log.Println("[agent] closing...")
			return
		case timestamp := <-healthcheck.C:
			healthcheckHandler(conn, timestamp)
		}
	}*/
}
