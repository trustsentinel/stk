package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/kr/pty"
	"golang.org/x/net/websocket"
)

func deserialise(buffer []byte) string {
	var r string
	err := json.Unmarshal(buffer[:], &r)
	if err != nil {
		fmt.Println("Can't deserislize", buffer[:])
	}

	fmt.Printf("%v => %v\n", buffer[:], r)
	output := append([]byte{byte(64)}, buffer[:]...)
	log.Println("->", output[:5], "...")
	return r
}

func incoming(r io.Reader, w io.Writer, S [32]byte) {
	buffer := make([]byte, 1024)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			panic(err)
		}
		d := Decrypt(buffer[:n], S)
		w.Write(d)
	}
}

func outgoing(r io.Reader, w io.Writer, S [32]byte) {
	buffer := make([]byte, 1024)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			panic(err)
		}
		e := Encrypt(buffer[:n], S)
		w.Write(e)
	}
}

// ShellServer offers connection
func ShellServer(S [32]byte) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		c := exec.Command("bash")
		f, err := pty.Start(c)
		if err != nil {
			ws.Write([]byte(fmt.Sprintf("Error creating pty: %s\r\n", err)))
			ws.Close()
			return
		}

		go incoming(ws, f, S)
		outgoing(f, ws, S)
		//go io.Copy(ws, f)
		//io.Copy(f, ws)
		ws.Close()
	}
}

func main() {
	//encrypted := false
	args := os.Args
	if len(args) > 2 {
		//encrypted, _ = strconv.ParseBool(args[2])
	}

	A := getRemoteKey()
	a := getPrivateRemoteKey() // testing purposal
	B, b := generatePair()
	S := getSharedKey(A, b)
	test(A, a, B, b, S)

	http.Handle("/", http.FileServer(http.Dir("./")))
	http.Handle("/channels/dcb6888d", websocket.Handler(ShellServer(S)))
	http.ListenAndServe(":12345", nil)
}
