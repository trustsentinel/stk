package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
	"websocket"
)

// authConnection .. set the auth connection to broker based on auth-websocket channel
func authConnection(hostname string, port int, path string) *websocket.Conn {
	address := url.URL{Scheme: "ws", Host: net.JoinHostPort(hostname, Int2string(port)), Path: path}
	headers := make(http.Header)
	connection, _, err := websocket.DefaultDialer.Dial(address.String(), headers)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return connection
}

// authHandler .. set the handler
func authHandler(token string, connection *websocket.Conn, shutdown chan bool) {
	for {
		messageType, data, err := connection.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		handleAuth(connection, messageType, data)
	}
}

// handleAuth .. deals with incoming messages
func handleAuth(connection *websocket.Conn, messageType int, packet []byte) {
	if messageType != websocket.BinaryMessage {
		return
	}
	// packet arrives...
}

// healthcheckHandler .. set the keep-alive handler
func healthcheckHandler(connection *websocket.Conn, timestamp time.Time) {
	if err := connection.WriteMessage(websocket.BinaryMessage, EncodeTimestamp(timestamp)); err != nil {
		log.Println(err)
		return
	}
}
