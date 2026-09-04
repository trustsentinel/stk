// Package crypto holds stk's small cryptographic and random helpers.
//
// It is the first piece ported to Go modules from the 2019 prototype; the
// transport layer (agent/auth) still depends on the defunct
// gopkg.in/noisesocket.v0 and is being refactored onto github.com/flynn/noise.
package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

// To32 copies a byte slice into a fixed 32-byte array (e.g. a Curve25519 key).
func To32(b []byte) [32]byte {
	var out [32]byte
	copy(out[:], b)
	return out
}

// Key32 decodes a standard-base64 string into a 32-byte key.
//
// NOTE: the original prototype's DecodeKey truncated to the first 4 bytes — a
// bug that silently produced a weak/invalid key. This copies the full key and
// reports decode errors.
func Key32(b64 string) ([32]byte, error) {
	var k [32]byte
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return k, err
	}
	copy(k[:], raw)
	return k, nil
}

// GenerateRandomBytes returns n securely-generated random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a URL-safe base64 string of n random bytes.
func GenerateRandomString(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
