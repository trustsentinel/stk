package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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
			log.Println("["+ID+"::broker] Incoming message from connected client", string(message))
			c := h.client
			err := c.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				log.Println("["+ID+"::broker] error data write:", err)
				break
			}

		case message := <-h.outgoing:
			log.Println("["+ID+"::broker] Outgoing message to connected client", string(message))
			h.commonIncoming <- message
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

// Allowed for 1000 concurrent channels
// theorical limitation is found below to 60k channels
// up to maximum allowed
var channels = make(map[string]chan bool, 1000)

/// createDedicatedChannel allow new upcoming messages into the specific pipeline
/// created by a new ws client-server connection under a randomly port (@todo not arbitrary?)
func createDedicatedChannel(channelID string, port int, c chan bool,
	incoming *chan []byte, outgoing *chan []byte,
	cin *chan []byte, cout *chan []byte) {

	log.Println("Creating channel ", channelID, "port:", port)

	channelMux := http.NewServeMux()
	channelMux.HandleFunc("/channel", goGet(channelID))
	channelMux.HandleFunc("/ws", secureWsChannel(channelID, incoming, outgoing, cin, cout))
	CreateDedicatedChannelResponder("localhost", port, channelMux)
	c <- true
}

func goGet(channelID string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("server belongs to " + channelID)
	}
}

func (h *channelHub) runCommon() {
	ID := "common"
	for {
		select {
		case c := <-h.newConnection:
			log.Printf("[%s::broker] New connection!\n", ID)
			h.client = c
		case message := <-h.commonIncoming:
			log.Println("["+ID+"::broker] Incoming message from secured!", string(message))
			c := h.client
			if c != nil {
				err := c.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Println("["+ID+"::broker] error data write:", err)
					break
				}
			} else {
				log.Println("["+ID+"::broker] no connection from client found", c)
			}

		case message := <-h.commonOutgoing:
			log.Println(`[`+ID+`] Outgoing message to secured channel`, string(message))
			h.incoming <- message
		}
	}
}

func goWsChannel(incoming *chan []byte, outgoing *chan []byte,
	cin *chan []byte, cout *chan []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		channelID := params["channel"]
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}
		fmt.Println("common ws -> channel address", incoming, *incoming)
		fmt.Println("handling channel", channelID)
		h := newPipelineHub(*incoming, *outgoing, *cin, *cout)
		commonWsHandler(ws, h)
	}
}

func commonWsHandler(c *websocket.Conn, h *channelHub) {
	go h.runCommon()

	h.newConnection <- c
	for {
		mt, message, _ := c.ReadMessage()
		log.Println("[common]: read from browser ", string(message), "forward to outoing message box", mt)
		h.commonOutgoing <- message
	}
}

func goMessageChannel(incoming *chan []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		channelID := params["channel"]
		messages, ok := r.URL.Query()["message"]

		if !ok || len(messages[0]) < 1 {
			log.Println("Url Param 'message' is missing")
			return
		}

		log.Println("Sending message to channel", channelID)
		*incoming <- []byte(messages[0])
		json.NewEncoder(w).Encode(channelID)
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
		if l > 20 {
			log.Printf(`[`+channelID+`::socket] read, size:`, l)
		} else {
			log.Printf(`[`+channelID+`::socket] read: `, string(message))
		}

		h.outgoing <- message
	}
}
