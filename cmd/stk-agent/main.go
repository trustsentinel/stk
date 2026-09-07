// Command stk-agent runs on the host being shelled into. It dials OUT to the hub
// (so it exposes no inbound port), completes a mutually-authenticated Noise
// handshake with the client, then brokers a PTY shell over the encrypted session.
package main

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"

	"github.com/trustsentinel/stk/internal/secure"
	"github.com/trustsentinel/stk/internal/shell"
	"github.com/trustsentinel/stk/internal/transport"
)

func main() {
	hub := flag.String("hub", "ws://localhost:8443/ws", "hub websocket URL")
	room := flag.String("room", "default", "rendezvous room")
	shellPath := flag.String("shell", "/bin/sh", "shell to spawn")
	priv := flag.String("key", "", "base64 static private key (generated if empty)")
	pub := flag.String("pubkey", "", "base64 static public key (with -key)")
	authClient := flag.String("authorized-client", "", "base64 client public key allowed to connect (empty = any)")
	once := flag.Bool("once", false, "serve a single session then exit")
	flag.Parse()

	kp, err := secure.LoadKeypair(*priv, *pub)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("stk-agent static pubkey: %s", secure.EncodePublic(kp.Public))

	var allowed [][]byte
	if *authClient != "" {
		pk, err := secure.DecodePublic(*authClient)
		if err != nil {
			log.Fatalf("bad -authorized-client: %v", err)
		}
		allowed = append(allowed, pk)
		log.Printf("authorizing only client %s", *authClient)
	} else {
		log.Print("WARNING: no -authorized-client set; accepting any authenticated client (demo only)")
	}

	for {
		if err := serve(*hub, *room, kp, allowed, *shellPath); err != nil {
			log.Printf("session ended: %v", err)
		}
		if *once {
			return
		}
		time.Sleep(time.Second)
	}
}

func serve(hub, room string, kp secure.Keypair, allowed [][]byte, shellPath string) error {
	u := hub + "?role=agent&room=" + url.QueryEscape(room)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return err
	}
	conn := transport.NewWSConn(c)
	log.Printf("connected to hub, waiting to be paired (room=%s)", room)

	sess, err := secure.Handshake(conn, secure.Config{Static: kp, Initiator: false, Authorized: allowed})
	if err != nil {
		conn.Close()
		return err
	}
	log.Printf("secure session established with client %s", secure.EncodePublic(sess.PeerStatic))

	sh, err := shell.Start(shellPath)
	if err != nil {
		sess.Close()
		return err
	}

	// PTY output -> encrypted session.
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, rerr := sh.Read(buf)
			if n > 0 {
				if werr := sess.Write(buf[:n]); werr != nil {
					break
				}
			}
			if rerr != nil {
				break
			}
		}
		sess.Close()
	}()

	// Encrypted session -> PTY input.
	for {
		data, rerr := sess.Read()
		if len(data) > 0 {
			if _, werr := sh.Write(data); werr != nil {
				break
			}
		}
		if rerr != nil {
			break
		}
	}
	sh.Close()
	sess.Close()
	return nil
}
