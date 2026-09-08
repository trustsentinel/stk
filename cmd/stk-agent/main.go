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
	identity := flag.String("identity", "", "path to a persistent device identity (created on first use)")
	priv := flag.String("key", "", "base64 static private key (with -pubkey; overridden by -identity)")
	pub := flag.String("pubkey", "", "base64 static public key (with -key)")
	authClient := flag.String("authorized-client", "", "base64 client public key allowed to connect")
	authClients := flag.String("authorized-clients", "", "path to an authorized-clients file (base64 keys, one per line; re-read each session)")
	once := flag.Bool("once", false, "serve a single session then exit")
	flag.Parse()

	kp, err := secure.ResolveIdentity(*identity, *priv, *pub)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("stk-agent static pubkey: %s", secure.EncodePublic(kp.Public))

	var staticAllowed [][]byte
	if *authClient != "" {
		pk, err := secure.DecodePublic(*authClient)
		if err != nil {
			log.Fatalf("bad -authorized-client: %v", err)
		}
		staticAllowed = append(staticAllowed, pk)
	}
	switch {
	case *authClients != "":
		log.Printf("authorizing clients from %s (re-read each session)", *authClients)
	case len(staticAllowed) > 0:
		log.Printf("authorizing only client %s", *authClient)
	default:
		log.Print("WARNING: no -authorized-client(s) set; accepting any authenticated client (demo only)")
	}

	// allowlist is rebuilt per session so newly enrolled keys take effect without
	// restarting the agent.
	allowlist := func() [][]byte {
		allowed := append([][]byte(nil), staticAllowed...)
		if *authClients != "" {
			fileKeys, ferr := secure.LoadAuthorizedKeys(*authClients)
			if ferr != nil {
				log.Printf("WARNING: reading %s: %v", *authClients, ferr)
			}
			allowed = append(allowed, fileKeys...)
		}
		return allowed
	}

	for {
		if err := serve(*hub, *room, kp, allowlist, *shellPath); err != nil {
			log.Printf("session ended: %v", err)
		}
		if *once {
			return
		}
		time.Sleep(time.Second)
	}
}

func serve(hub, room string, kp secure.Keypair, allowlist func() [][]byte, shellPath string) error {
	u := hub + "?role=agent&room=" + url.QueryEscape(room)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return err
	}
	conn := transport.NewWSConn(c)
	log.Printf("connected to hub, waiting to be paired (room=%s)", room)

	sess, err := secure.Handshake(conn, secure.Config{Static: kp, Initiator: false, Authorized: allowlist()})
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
