package main

import (
	"io"
	"log"
	"os"
)

func wrap(r io.Reader, w io.Writer) {
	buffer := make([]byte, 1024)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			panic(err)
		}
		output := append([]byte{byte(n)}, buffer[:n]...)
		log.Println("read ", n, "->", string(output))
		w.Write(output)
	}
}

func main2() {

	//r, w, _ := os.Pipe()
	wrap(os.Stdin, os.Stdout)
	//go io.Copy(w, os.Stdin)
	//io.Copy(os.Stdout, r)
	//io.Copy(f, ws)
}
