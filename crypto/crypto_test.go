package crypto

import (
	"encoding/base64"
	"testing"
)

func TestKey32RoundTrip(t *testing.T) {
	var raw [32]byte
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	k, err := Key32(base64.StdEncoding.EncodeToString(raw[:]))
	if err != nil {
		t.Fatalf("Key32: %v", err)
	}
	if k != raw {
		t.Fatalf("round-trip mismatch")
	}
}

func TestKey32FullLength(t *testing.T) {
	// Regression: the original prototype truncated to 4 bytes.
	b := make([]byte, 32)
	for i := range b {
		b[i] = 0xAB
	}
	k, _ := Key32(base64.StdEncoding.EncodeToString(b))
	for i, v := range k {
		if v != 0xAB {
			t.Fatalf("byte %d not copied (truncation bug regressed): %#x", i, v)
		}
	}
}

func TestKey32BadInput(t *testing.T) {
	if _, err := Key32("!!! not base64 !!!"); err == nil {
		t.Fatal("expected error on invalid base64")
	}
}

func TestGenerateRandomBytesLen(t *testing.T) {
	b, err := GenerateRandomBytes(16)
	if err != nil || len(b) != 16 {
		t.Fatalf("len=%d err=%v", len(b), err)
	}
}

func TestGenerateRandomStringUnique(t *testing.T) {
	a, _ := GenerateRandomString(24)
	b, _ := GenerateRandomString(24)
	if a == "" || a == b {
		t.Fatalf("expected non-empty unique strings: %q / %q", a, b)
	}
}

func TestTo32(t *testing.T) {
	g := To32([]byte{1, 2, 3})
	if g[0] != 1 || g[1] != 2 || g[2] != 3 || g[3] != 0 {
		t.Fatalf("To32 wrong: %v", g)
	}
}
