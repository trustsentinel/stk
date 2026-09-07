package secure_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/trustsentinel/stk/internal/secure"
	"github.com/trustsentinel/stk/internal/transport"
)

// handshakePair runs an initiator and responder concurrently over an in-memory
// pipe and returns both sessions (and any errors).
func handshakePair(t *testing.T, iCfg, rCfg secure.Config) (*secure.Session, error, *secure.Session, error) {
	t.Helper()
	ci, cr := transport.Pipe()

	type res struct {
		s   *secure.Session
		err error
	}
	rch := make(chan res, 1)
	go func() {
		s, err := secure.Handshake(cr, rCfg)
		rch <- res{s, err}
	}()
	is, ierr := secure.Handshake(ci, iCfg)
	r := <-rch
	return is, ierr, r.s, r.err
}

func TestHandshakeMutualAuthAndExchange(t *testing.T) {
	client, _ := secure.GenerateKeypair()
	agent, _ := secure.GenerateKeypair()

	cs, cerr, ss, serr := handshakePair(t,
		secure.Config{Static: client, Initiator: true, Authorized: [][]byte{agent.Public}},
		secure.Config{Static: agent, Initiator: false, Authorized: [][]byte{client.Public}},
	)
	if cerr != nil {
		t.Fatalf("client handshake: %v", cerr)
	}
	if serr != nil {
		t.Fatalf("agent handshake: %v", serr)
	}

	if !bytes.Equal(cs.PeerStatic, agent.Public) {
		t.Error("client authenticated the wrong agent key")
	}
	if !bytes.Equal(ss.PeerStatic, client.Public) {
		t.Error("agent authenticated the wrong client key")
	}

	// Exchange in both directions over the encrypted session.
	if err := cs.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	got, err := ss.Read()
	if err != nil || string(got) != "ping" {
		t.Fatalf("agent read %q, %v", got, err)
	}
	if err := ss.Write([]byte("pong")); err != nil {
		t.Fatal(err)
	}
	got, err = cs.Read()
	if err != nil || string(got) != "pong" {
		t.Fatalf("client read %q, %v", got, err)
	}
}

func TestUnauthorizedClientRejectedByAgent(t *testing.T) {
	client, _ := secure.GenerateKeypair()
	agent, _ := secure.GenerateKeypair()
	stranger, _ := secure.GenerateKeypair()

	// The agent (responder) only authorizes `stranger`, so the real client's key
	// must be rejected — this is the gate that stops unauthorized shell access.
	_, _, _, serr := handshakePair(t,
		secure.Config{Static: client, Initiator: true, Authorized: [][]byte{agent.Public}},
		secure.Config{Static: agent, Initiator: false, Authorized: [][]byte{stranger.Public}},
	)
	if serr == nil {
		t.Fatal("expected the agent to reject an unauthorized client")
	}
	if !errors.Is(serr, secure.ErrUnauthorized) {
		t.Fatalf("want ErrUnauthorized, got %v", serr)
	}
}

func TestEmptyAllowlistAcceptsAnyAuthenticatedPeer(t *testing.T) {
	client, _ := secure.GenerateKeypair()
	agent, _ := secure.GenerateKeypair()

	_, cerr, _, serr := handshakePair(t,
		secure.Config{Static: client, Initiator: true}, // no Authorized set
		secure.Config{Static: agent, Initiator: false}, // no Authorized set
	)
	if cerr != nil || serr != nil {
		t.Fatalf("empty allow-list should accept: client=%v agent=%v", cerr, serr)
	}
}
