package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"websocket"

	"github.com/flynn/noise"
	"gopkg.in/noisesocket.v0"
)

type channelHub struct {
	client         *websocket.Conn
	newConnection  chan *websocket.Conn
	commonIncoming chan []byte
	commonOutgoing chan []byte
	incoming       chan []byte
	outgoing       chan []byte
}

func (h *channelHub) run(ID string) {
	for {
		select {
		case conn := <-h.newConnection:
			log.Println(`[` + ID + `] New connection!`)
			h.client = conn
		case message := <-h.incoming:
			log.Println("INCOMING", message)
			c := h.client

			e := message[:]
			fmt.Println("[out] writing plain... ", e, string(e))
			if encrypted {
				e = Encrypt(message[:], sharedKey)
				fmt.Println("[out] writing encripted... ", e, string(e))
			}

			err := c.WriteMessage(websocket.BinaryMessage, e)
			if err != nil {
				log.Println("["+ID+"::broker] error data write:", err)
				break
			}

		case message := <-h.outgoing:
			log.Println("OUTGOING", message)
			//log.Println("["+ID+"::broker] Outgoing message to connected client", string(message))
			//h.commonIncoming <- message
			// testing purposal, echoing messsages with length payload
			//h.outgoing <- appendPayload(message)
		}
	}
}

func newPipelineHub(incoming chan []byte, outgoing chan []byte, cincoming chan []byte, coutgoing chan []byte) *channelHub {
	fmt.Println("newHub -> channel address", incoming)
	log.Printf("cin %v", cincoming)
	log.Printf("cin %v", coutgoing)
	log.Printf("in %v", incoming)
	log.Printf("out %v", outgoing)
	return &channelHub{
		client:         nil,
		newConnection:  make(chan *websocket.Conn, 1),
		incoming:       incoming,
		outgoing:       outgoing,
		commonIncoming: cincoming,
		commonOutgoing: coutgoing,
	}
}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handlerSecureChannel(channelID string, c *websocket.Conn, h *channelHub) {
	go h.run(channelID)
	h.newConnection <- c
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println(`[`+channelID+`]  error reading:`, err)
			break
		}
		l := len(message)
		if l > 70 {
			log.Printf(`[`+channelID+`::socket] read, size: %d`, l)
		} else {
			log.Printf(`[`+channelID+`::socket] read: %s`, string(message))
		}
		d := message[:]
		fmt.Println("[in] reading encrypted... ", d, string(d))
		if encrypted {
			d = Decrypt(message[:], sharedKey)
			fmt.Println("[in] deciphering... ", d, string(d))
		}
		h.outgoing <- d
	}
}
func secureWsChannel(channelID string, incoming *chan []byte, outgoing *chan []byte, cin *chan []byte, cout *chan []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		h := newPipelineHub(*incoming, *outgoing, *cin, *cout)
		handlerSecureChannel(channelID, ws, h)
	}
}

var cin = make(chan []byte, 1)
var cout = make(chan []byte, 1)
var incoming = make(chan []byte, 1)
var outgoing = make(chan []byte, 1)

func read(r io.Reader) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scan := bufio.NewScanner(r)
		for scan.Scan() {
			s := scan.Text()
			lines <- s
		}
	}()
	return lines
}

var encrypted = true
var sharedKey = [32]byte{16, 44, 122, 53, 10, 77, 214, 164, 220, 19, 151, 64, 250, 13, 191, 5, 61, 6, 88, 138, 89, 1, 190, 18, 119, 176, 132, 39, 227, 233, 73, 35}

func main() {
	if !encrypted {
		log.Println("*******************************************\nENCRYPTED IS DISABLED\n*******************************************")
	}
	channelID := "aaaaa"
	channelMux := http.NewServeMux()
	channelMux.HandleFunc("/ws", secureWsChannel(channelID, &incoming, &outgoing, &cin, &cout))
	go CreateDedicatedChannelResponder("localhost", 8100, channelMux)

	input := make(chan []byte)

	go func(in chan []byte) {
		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				close(in)
				log.Println("Error in read string", err)
				panic(err)
			}
			in <- []byte(s)
		}
	}(input)

	for {
		select {
		case in := <-input:
			fmt.Println("Read from stdin: ", string(in))
			log.Printf("in %v", incoming)
			incoming <- in

		case msg := <-outgoing:
			fmt.Println(">" + string(msg))
		}
	}
}
