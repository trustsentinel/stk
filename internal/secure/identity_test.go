package secure_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/trustsentinel/stk/internal/secure"
)

func TestLoadOrCreateIdentityPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "id")

	kp1, err := secure.LoadOrCreateIdentity(path) // creates
	if err != nil {
		t.Fatal(err)
	}
	if len(kp1.Public) == 0 || len(kp1.Private) == 0 {
		t.Fatal("empty keypair created")
	}

	kp2, err := secure.LoadOrCreateIdentity(path) // loads the same
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(kp1.Public, kp2.Public) || !bytes.Equal(kp1.Private, kp2.Private) {
		t.Fatal("identity not stable across loads")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("identity file mode = %o, want 600", perm)
	}
}

func TestLoadAuthorizedKeys(t *testing.T) {
	a, _ := secure.GenerateKeypair()
	b, _ := secure.GenerateKeypair()
	path := filepath.Join(t.TempDir(), "authorized_clients")
	content := "# team keys\n\n" +
		secure.EncodePublic(a.Public) + "   alice@laptop\n" +
		secure.EncodePublic(b.Public) + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	keys, err := secure.LoadAuthorizedKeys(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(keys))
	}
	if !bytes.Equal(keys[0], a.Public) || !bytes.Equal(keys[1], b.Public) {
		t.Error("parsed keys do not match (comment/label handling wrong)")
	}
}

func TestLoadAuthorizedKeysRejectsGarbage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad")
	if err := os.WriteFile(path, []byte("not-valid-base64!!!\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := secure.LoadAuthorizedKeys(path); err == nil {
		t.Fatal("expected an error for invalid key material")
	}
}
