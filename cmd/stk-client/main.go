// Command stk-client connects to the hub, completes a mutually-authenticated
// Noise handshake with the agent (the hub only relays ciphertext), and drives the
// remote shell. With -exec it runs one command and exits (used by the e2e test);
// otherwise it is interactive, wiring stdin/stdout to the remote PTY.
package main

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"

	"github.com/trustsentinel/stk/internal/secure"
	"github.com/trustsentinel/stk/internal/transport"
)

func main() {
	hub := flag.String("hub", "ws://localhost:8443/ws", "hub websocket URL")
	room := flag.String("room", "default", "rendezvous room")
	identity := flag.String("identity", "", "path to a persistent device identity (created on first use)")
	priv := flag.String("key", "", "base64 static private key (with -pubkey; overridden by -identity)")
	pub := flag.String("pubkey", "", "base64 static public key (with -key)")
	authAgent := flag.String("authorized-agent", "", "base64 agent public key to pin (required)")
	execCmd := flag.String("exec", "", "run one command then exit (non-interactive)")
	flag.Parse()

	kp, err := secure.ResolveIdentity(*identity, *priv, *pub)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("stk-client static pubkey: %s", secure.EncodePublic(kp.Public))

	if *authAgent == "" {
		log.Fatal("-authorized-agent (the agent's base64 public key) is required: the client pins the agent it will talk to")
	}
	pin, err := secure.DecodePublic(*authAgent)
	if err != nil {
		log.Fatalf("bad -authorized-agent: %v", err)
	}

	u := *hub + "?role=client&room=" + url.QueryEscape(*room)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatalf("dial hub: %v", err)
	}
	conn := transport.NewWSConn(c)

	sess, err := secure.Handshake(conn, secure.Config{Static: kp, Initiator: true, PeerStatic: pin})
	if err != nil {
		log.Fatalf("handshake/auth failed: %v", err)
	}
	log.Printf("secure session established with agent %s", secure.EncodePublic(sess.PeerStatic))

	if *execCmd != "" {
		if err := sess.Write([]byte(*execCmd + "; exit\n")); err != nil {
			log.Fatalf("write: %v", err)
		}
		for {
			data, rerr := sess.Read()
			if len(data) > 0 {
				os.Stdout.Write(data)
			}
			if rerr != nil {
				break
			}
		}
		return
	}

	// Interactive: stdin -> session, session -> stdout.
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, rerr := os.Stdin.Read(buf)
			if n > 0 {
				if werr := sess.Write(buf[:n]); werr != nil {
					break
				}
			}
			if rerr != nil {
				break
			}
		}
	}()
	for {
		data, rerr := sess.Read()
		if len(data) > 0 {
			os.Stdout.Write(data)
		}
		if rerr != nil {
			if rerr != io.EOF {
				log.Printf("session closed: %v", rerr)
			}
			break
		}
	}
}
