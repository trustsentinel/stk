package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/flynn/noise"
	"gopkg.in/noisesocket.v0"
)

func CreateDedicatedChannelResponder(address string, port int, channelMux *http.ServeMux) {

	pub, _ := base64.StdEncoding.DecodeString("J6TRfRXR5skWt6w5cFyaBxX8LPeIVxboZTLXTMhk4HM=")
	priv, _ := base64.StdEncoding.DecodeString("vFilCT/FcyeShgbpTUrpru9n5yzZey8yfhsAx6DeL80=")

	fmt.Println("server ss.pub", pub)
	fmt.Println("server ss.priv", priv)
	serverKeys := noise.DHKey{
		Public:  pub,
		Private: priv,
	}

	listener, err := noisesocket.Listen(":"+strconv.FormatInt(int64(port), 10),
		&noisesocket.ConnectionConfig{StaticKey: serverKeys})
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}

	if err := http.Serve(listener, channelMux); err != nil {
		log.Fatal(err)
	}
}
