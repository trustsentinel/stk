// Package secure establishes an end-to-end encrypted, mutually authenticated
// channel over a transport.MsgConn using the Noise Protocol (IK pattern).
//
// The hub only relays these messages; it never holds the keys, so it cannot read
// the session — that end-to-end property is stk's whole point.
//
// IK is chosen deliberately for a broker with anonymous rendezvous:
//   - The initiator (client) must already know the responder's (agent's) static
//     key, so it authenticates the agent up front — no trust in the hub, and no
//     MITM. That key is a required input, not an after-the-fact check.
//   - The initiator's static key is sent in the FIRST message, so the responder
//     authenticates the client immediately and drops an unauthorized client
//     before completing the handshake or spawning a shell (XX only learned it on
//     the third message, after the session was half-established).
package secure

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// LoadOrCreateIdentity returns the persistent device identity stored at path,
// creating (and persisting, mode 0600) a fresh keypair on first use. The file is
// two base64 lines: private key, then public key. This gives each device a stable
// long-lived identity instead of a per-process ephemeral key.
func LoadOrCreateIdentity(path string) (Keypair, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		fields := strings.Fields(string(data))
		if len(fields) < 2 {
			return Keypair{}, fmt.Errorf("secure: identity file %s is malformed", path)
		}
		priv, perr := base64.StdEncoding.DecodeString(fields[0])
		pub, uerr := base64.StdEncoding.DecodeString(fields[1])
		if perr != nil || uerr != nil {
			return Keypair{}, fmt.Errorf("secure: identity file %s has invalid base64", path)
		}
		return Keypair{Public: pub, Private: priv}, nil
	}
	if !os.IsNotExist(err) {
		return Keypair{}, err
	}
	kp, err := GenerateKeypair()
	if err != nil {
		return Keypair{}, err
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return Keypair{}, err
		}
	}
	contents := EncodePublic(kp.Private) + "\n" + EncodePublic(kp.Public) + "\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		return Keypair{}, err
	}
	return kp, nil
}

// ResolveIdentity picks a static key: a persistent identity file if identityPath
// is set (created on first use), else an explicit base64 -key/-pubkey pair, else a
// fresh ephemeral key. Shared by the agent and client.
func ResolveIdentity(identityPath, privB64, pubB64 string) (Keypair, error) {
	if identityPath != "" {
		return LoadOrCreateIdentity(identityPath)
	}
	return LoadKeypair(privB64, pubB64)
}

// LoadAuthorizedKeys parses an authorized-clients file: one base64 public key per
// line, "#" comments and blank lines ignored, any trailing text after the key
// (a label) ignored — the SSH authorized_keys convention. Re-read per session so
// enrolling a new client takes effect without restarting the agent.
func LoadAuthorizedKeys(path string) ([][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var keys [][]byte
	sc := bufio.NewScanner(f)
	for line := 0; sc.Scan(); line++ {
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		k, derr := base64.StdEncoding.DecodeString(strings.Fields(text)[0])
		if derr != nil {
			return nil, fmt.Errorf("secure: %s line %d: invalid key: %w", path, line+1, derr)
		}
		keys = append(keys, k)
	}
	return keys, sc.Err()
}

// Config configures a handshake.
type Config struct {
	Static    Keypair
	Initiator bool
	// PeerStatic is the responder's static public key. REQUIRED for the initiator
	// (IK pins the responder up front); ignored for the responder.
	PeerStatic []byte
	// Authorized, if non-empty, is the set of peer static public keys allowed to
	// complete the handshake (checked by the responder against the initiator's
	// key). Empty means "accept any authenticated peer" (demo/dev only).
	Authorized [][]byte
}

// ErrUnauthorized is returned when the peer's static key is not in Authorized.
var ErrUnauthorized = errors.New("secure: peer key not authorized")

// ErrNoPeerStatic is returned when an initiator handshake is attempted without
// pinning the responder's static key (required by IK).
var ErrNoPeerStatic = errors.New("secure: initiator must pin the responder's static key (PeerStatic)")

// Session is an established encrypted channel over a MsgConn.
type Session struct {
	conn transport.MsgConn
	send *noise.CipherState
	recv *noise.CipherState
	// PeerStatic is the remote party's authenticated static public key.
	PeerStatic []byte
}

// Handshake performs a Noise IK handshake over conn and returns an encrypted
// Session. The initiator pins the responder via cfg.PeerStatic; the responder
// authenticates the initiator against cfg.Authorized on the first message and
// drops an unauthorized peer before completing the handshake.
func Handshake(conn transport.MsgConn, cfg Config) (*Session, error) {
	if cfg.Initiator && len(cfg.PeerStatic) == 0 {
		return nil, ErrNoPeerStatic
	}
	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:   cipherSuite,
		Random:        rand.Reader,
		Pattern:       noise.HandshakeIK,
		Initiator:     cfg.Initiator,
		StaticKeypair: noise.DHKey{Public: cfg.Static.Public, Private: cfg.Static.Private},
		PeerStatic:    cfg.PeerStatic, // used by the initiator; empty for the responder
	})
	if err != nil {
		return nil, err
	}

	// IK is a 2-message handshake:
	//   1. initiator -> responder:  e, es, s, ss   (carries the initiator's static)
	//   2. responder -> initiator:  e, ee, se
	// The (send, recv) CipherState pair is returned on the second message.
	var c0, c1 *noise.CipherState
	if cfg.Initiator {
		out, _, _, werr := hs.WriteMessage(nil, nil) // msg1
		if werr != nil {
			return nil, werr
		}
		if err = conn.WriteMsg(out); err != nil {
			return nil, err
		}
		msg2, rerr := conn.ReadMsg()
		if rerr != nil {
			return nil, rerr
		}
		if _, c0, c1, err = hs.ReadMessage(nil, msg2); err != nil { // completes
			return nil, err
		}
	} else {
		msg1, rerr := conn.ReadMsg()
		if rerr != nil {
			return nil, rerr
		}
		if _, _, _, err = hs.ReadMessage(nil, msg1); err != nil {
			return nil, err
		}
		// Authenticate the initiator NOW, before completing the handshake.
		peer := hs.PeerStatic()
		if !authorized(peer, cfg.Authorized) {
			conn.Close()
			return nil, fmt.Errorf("%w: %s", ErrUnauthorized, EncodePublic(peer))
		}
		out, cc0, cc1, werr := hs.WriteMessage(nil, nil) // msg2, completes
		if werr != nil {
			return nil, werr
		}
		if err = conn.WriteMsg(out); err != nil {
			return nil, err
		}
		c0, c1 = cc0, cc1
	}
	if c0 == nil || c1 == nil {
		return nil, errors.New("secure: handshake did not complete")
	}

	s := &Session{conn: conn, PeerStatic: hs.PeerStatic()}
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
