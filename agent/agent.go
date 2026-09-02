package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var keys = make(map[int32][32]byte)
var S [32]byte

// authorativeHandler handle all incoming data from connection and handle under
// handleAuthPacket
func authorativeHandler(token string, connection *websocket.Conn, shutdown chan bool, localPairKeys *Keys) {
	for {
		messageType, data, err := connection.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("[auth]: packet received %x \n", data)
		handleAuthPacket(connection, messageType, data, localPairKeys)
	}
}

// handleAuthPacket receives and perform actions based on requests received from connection
func handleAuthPacket(connection *websocket.Conn, messageType int, packet []byte, localPairKeys *Keys) {
	log.Println("[auth] packet type: ", messageType, packet)

	if messageType != websocket.BinaryMessage {
		// ignoring text based messages
		return
	}

	req := GetRequest(packet)
	switch req.GetId() {
	case 1:
		// Auth reception means a user want to claim for authentication over device and key exchange
		// responding him back with a AuthAck message
		auth := DeserialiseAuth(packet)
		fmt.Printf("[auth] received command 'auth' from sid= %d, key=%s\n", auth.GetSid(), auth.GetPublicKey())

		// @todo: device key should be generated on each session
		dpk, _ := GetDeviceKey()

		// @todo: it should defines if user is authorised to perform
		// a connection under this device
		authorise(auth)

		// builds the response
		packet := SerialiseAuthAck(auth.GetSid(), base64.StdEncoding.EncodeToString(dpk))
		if err := connection.WriteMessage(messageType, packet); err != nil {
			log.Println(err)
			return
		}
	case 4:
		// AuthAck contains the information of channel created by authorative broker entity
		ack := DeserialiseAuthAck(packet)
		fmt.Println("[auth] received auth ack 2, secure channel created under port:", ack.GetPort())
		A := to32([]byte{16, 44, 122, 53, 10, 77, 214, 164, 220, 19, 151, 64, 250, 13, 191, 5, 61, 6, 88, 138, 89, 1, 190, 18, 119, 176, 132, 39, 227, 233, 73, 35})
		go CreateDedicatedChannel(ack.GetAddress(), ack.GetPort(), &A, localPairKeys)
	default:
		fmt.Println("[auth] unknown command")
		log.Println("[auth] sending message back", packet)
		if err := connection.WriteMessage(messageType, packet); err != nil {
			log.Println(err)
			return
		}
	}
}

func handleHeartbeat(connection *websocket.Conn, timestamp time.Time) {
	log.Println("[auth] Sending heartbeat, timestamp:", timestamp)
	/*if err := connection.WriteMessage(websocket.BinaryMessage, EncodeTimestamp(timestamp)); err != nil {
		log.Println(err)
		return
	}*/
}

// AuthConnection conforms the connection using a websocket client thru
// the provided path
func AuthConnection(token string, config *configuration) *websocket.Conn {
	hostname := net.JoinHostPort(config.Hostname, int2string(config.Port))
	address := url.URL{Scheme: "ws", Host: hostname, Path: "/devices/" + token}
	log.Printf("[agent] Connecting to authorative server under. %s", address.String())

	// header helps authorative server to identify roles
	// @todo. it should encrypted on some term or send hashed
	headers := make(http.Header)
	headers.Add("Stk-role", "agent")
	connection, _, err := websocket.DefaultDialer.Dial(address.String(), headers)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return connection
}

var encryptedEnabled = true

// pair of keys generated
var pair *Keys

func main() {
	pair := &Keys{}
	pair.generate()
	fmt.Println(pair.private)

	t := AgentToken()
	var token = flag.String("token", t, "specific token instead of generate or using saved one")
	encryptedEnabled = *flag.Bool("disableEncryption", true, "disable encryption")
	flag.Parse()
	log.SetFlags(0) // redirect output into logger

	if encryptedEnabled {
		log.Println("Encryption: enabled")
	} else {
		log.Println("Encryption: disabled")
	}

	// retrieve configuration hostname:port
	// of authorative server
	var c configuration
	c.getConfiguration()

	// catch the ctr+c signal to interrupt
	// the functionality
	interrupt := CloseIfInterrupt()

	// it allows gracefull restarting
	shutdown := make(chan bool)

	// it creates the connection to authorative server
	// using the generated token
	connection := AuthConnection(*token, &c)
	defer connection.Close()

	// main handler
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		authorativeHandler(*token, connection, shutdown, pair)
	}()

	// builds an interval signal used as Hearbeat
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-shutdown:
			log.Println("[agent] Closing...")
			return
		case <-finished:
			return
		case timestamp := <-heartbeat.C:
			handleHeartbeat(connection, timestamp)
		case <-interrupt:
			err := connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-finished:
			}
			return
		}
	}
}
