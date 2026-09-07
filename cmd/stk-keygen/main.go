// Command stk-keygen writes fresh Noise static keypairs to a directory, one pair
// per name given: <dir>/<name>.key (base64 private) and <dir>/<name>.pub (base64
// public). Used by the compose demo to provision agent and client identities
// before they start, so the demo authenticates with generated keys, not
// hardcoded ones.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/trustsentinel/stk/internal/secure"
)

func main() {
	dir := flag.String("dir", ".", "output directory")
	flag.Parse()
	names := flag.Args()
	if len(names) == 0 {
		log.Fatal("usage: stk-keygen -dir DIR NAME [NAME...]")
	}
	if err := os.MkdirAll(*dir, 0o700); err != nil {
		log.Fatal(err)
	}
	for _, name := range names {
		kp, err := secure.GenerateKeypair()
		if err != nil {
			log.Fatal(err)
		}
		writeFile(filepath.Join(*dir, name+".key"), secure.EncodePublic(kp.Private), 0o600)
		writeFile(filepath.Join(*dir, name+".pub"), secure.EncodePublic(kp.Public), 0o644)
		log.Printf("wrote %s.key / %s.pub (pub=%s)", name, name, secure.EncodePublic(kp.Public))
	}
}

func writeFile(path, contents string, mode os.FileMode) {
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		log.Fatal(err)
	}
}
