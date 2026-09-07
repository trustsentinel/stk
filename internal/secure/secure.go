// Package secure establishes an end-to-end encrypted, mutually authenticated
// channel over a transport.MsgConn using the Noise Protocol (XX pattern).
//
// The hub only relays these messages; it never holds the keys, so it cannot read
// the session — that end-to-end property is stk's whole point. XX gives mutual
// static-key authentication: after the handshake each side learns the other's
// static public key, which is checked against an allow-list.
package secure

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/flynn/noise"

	"github.com/trustsentinel/stk/internal/transport"
)

// cipherSuite: Curve25519 + ChaCha20-Poly1305 + BLAKE2b.
var cipherSuite = noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2b)

// Keypair is a Noise static keypair (Curve25519).
type Keypair struct {
	Public  []byte
	Private []byte
}

// GenerateKeypair returns a fresh static keypair.
func GenerateKeypair() (Keypair, error) {
	dh, err := cipherSuite.GenerateKeypair(rand.Reader)
	if err != nil {
		return Keypair{}, err
	}
	return Keypair{Public: dh.Public, Private: dh.Private}, nil
}

// LoadKeypair decodes a base64 private+public keypair; if privB64 is empty it
// generates a fresh one.
func LoadKeypair(privB64, pubB64 string) (Keypair, error) {
	if privB64 == "" {
		return GenerateKeypair()
	}
	priv, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil {
		return Keypair{}, fmt.Errorf("secure: bad private key: %w", err)
	}
	pub, err := base64.StdEncoding.DecodeString(pubB64)
	if err != nil {
		return Keypair{}, fmt.Errorf("secure: bad public key: %w", err)
	}
	return Keypair{Public: pub, Private: priv}, nil
}

// EncodePublic returns the base64 form of a public key.
func EncodePublic(pub []byte) string { return base64.StdEncoding.EncodeToString(pub) }

// DecodePublic parses a base64 public key.
func DecodePublic(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }

// Config configures a handshake.
type Config struct {
	Static    Keypair
	Initiator bool
	// Authorized, if non-empty, is the set of peer static public keys allowed to
	// complete the handshake. Empty means "accept any" (demo/dev only).
	Authorized [][]byte
}

// ErrUnauthorized is returned when the peer's static key is not in Authorized.
var ErrUnauthorized = errors.New("secure: peer key not authorized")

// Session is an established encrypted channel over a MsgConn.
type Session struct {
	conn transport.MsgConn
	send *noise.CipherState
	recv *noise.CipherState
	// PeerStatic is the remote party's authenticated static public key.
	PeerStatic []byte
}

// Handshake performs a Noise XX handshake over conn and returns an encrypted
// Session, rejecting the peer if its static key is not authorized.
func Handshake(conn transport.MsgConn, cfg Config) (*Session, error) {
	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:   cipherSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeXX,
		Initiator:     cfg.Initiator,
		StaticKeypair: noise.DHKey{Public: cfg.Static.Public, Private: cfg.Static.Private},
	})
	if err != nil {
		return nil, err
	}

	// XX is a 3-message handshake:
	//   1. initiator -> responder:  e
	//   2. responder -> initiator:  e, ee, s, es
	//   3. initiator -> responder:  s, se
	// The initiator writes messages 1 and 3; the responder writes message 2.
	// The (send, recv) CipherState pair is returned on the final message.
	var c0, c1 *noise.CipherState
	for step := 1; step <= 3; step++ {
		iWrite := (step%2 == 1) == cfg.Initiator
		if iWrite {
			var out []byte
			out, c0, c1, err = hs.WriteMessage(nil, nil)
			if err != nil {
				return nil, err
			}
			if err = conn.WriteMsg(out); err != nil {
				return nil, err
			}
		} else {
			msg, rerr := conn.ReadMsg()
			if rerr != nil {
				return nil, rerr
			}
			if _, c0, c1, err = hs.ReadMessage(nil, msg); err != nil {
				return nil, err
			}
		}
	}
	if c0 == nil || c1 == nil {
		return nil, errors.New("secure: handshake did not complete")
	}

	peer := hs.PeerStatic()
	if !authorized(peer, cfg.Authorized) {
		conn.Close()
		return nil, fmt.Errorf("%w: %s", ErrUnauthorized, EncodePublic(peer))
	}

	s := &Session{conn: conn, PeerStatic: peer}
	// c0 encrypts initiator->responder, c1 encrypts responder->initiator.
	if cfg.Initiator {
		s.send, s.recv = c0, c1
	} else {
		s.send, s.recv = c1, c0
	}
	return s, nil
}

// Write encrypts p and sends it as one message.
func (s *Session) Write(p []byte) error {
	ct, err := s.send.Encrypt(nil, nil, p)
	if err != nil {
		return err
	}
	return s.conn.WriteMsg(ct)
}

// Read receives one message and decrypts it.
func (s *Session) Read() ([]byte, error) {
	ct, err := s.conn.ReadMsg()
	if err != nil {
		return nil, err
	}
	return s.recv.Decrypt(nil, nil, ct)
}

// Close closes the underlying connection.
func (s *Session) Close() error { return s.conn.Close() }

func authorized(peer []byte, allowed [][]byte) bool {
	if len(allowed) == 0 {
		return true // demo/dev: accept any authenticated peer
	}
	for _, a := range allowed {
		if len(a) == len(peer) && subtle.ConstantTimeCompare(a, peer) == 1 {
			return true
		}
	}
	return false
}
