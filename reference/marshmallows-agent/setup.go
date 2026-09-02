package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	rand "math/rand"
	"net"
	"net/http"
	"os"
	"time"

	requests "./protocol"
	"github.com/flynn/noise"
	"github.com/golang/protobuf/proto"
	"gopkg.in/noisesocket.v0"
)

func readToken() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please, enter agent token: ")
	token, _ := reader.ReadString('\n')
	return token
}

// Generate ID
func generateId() string {
	var charset = "0123456789abcdef"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var length = 6
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// sendAuthentication ... sends the authentication request to broker
func sendAuthentication(address string, secretToken string, keys *noise.DHKey, deviceID string, handler func([]byte)) {
	publicKey := base64.StdEncoding.EncodeToString(keys.Public)
	auth := &requests.AgentAuth{
		Token:     proto.String(secretToken),
		PublicKey: proto.String(publicKey),
		DeviceId:  proto.String(deviceID),
	}
	data, err := proto.Marshal(auth)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	secureHTTPPost(address, keys, data, handler)
}

func secureHTTPGet(address string, keys *noise.DHKey) {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DisableKeepAlives:   true,

		DialTLS: func(network, addr string) (net.Conn, error) {
			conn, err := noisesocket.Dial(addr, &noisesocket.ConnectionConfig{StaticKey: *keys})
			if err != nil {
				fmt.Println("Dial", err)
			}
			return conn, err
		},
	}

	c := make(chan bool)
	go func() {
		cli := &http.Client{
			Transport: transport,
		}
		buffer := make([]byte, 1024)
		reader := bytes.NewReader(buffer)
		req, err := http.NewRequest("GET", "https://"+address, reader)
		if err != nil {
			panic(err)
		}

		resp, err := cli.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			panic(err)
		}
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
		c <- true
	}()

	<-c
	fmt.Println("done.")
}

// secureHTTPPost ... sends a Http Post protected with a Noise channel
func secureHTTPPost(address string, keys *noise.DHKey, buffer []byte, handleResponse func([]byte)) {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DisableKeepAlives:   true,

		DialTLS: func(network, addr string) (net.Conn, error) {
			conn, err := noisesocket.Dial(addr, &noisesocket.ConnectionConfig{StaticKey: *keys})
			if err != nil {
				fmt.Println("Dial", err)
			}
			return conn, err
		},
	}

	c := make(chan bool)
	go func() {
		cli := &http.Client{
			Transport: transport,
		}
		reader := bytes.NewReader(buffer)
		req, err := http.NewRequest("POST", "https://"+address, reader)
		if err != nil {
			panic(err)
		}

		resp, err := cli.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}

		response, err := ioutil.ReadAll(resp.Body)
		//_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			panic(err)
		}
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
		fmt.Printf("RESPONSE %s", response)
		if response != nil {
			handleResponse(response)
		}

		c <- true
	}()

	<-c
	fmt.Println("done.")
}
