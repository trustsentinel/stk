package main

import (
	"fmt"

	"log"

	"github.com/gorilla/websocket"
)

// Client wrapper of connection associated to
// one client connections from hosts under registration
type Client struct {
	*Peer
	token     string
	sid       int32
	timestamp int64
}

func (client *Client) close() {

	// Close the current connection
	client.Peer.close()

	// Disconnects peer from peers pool
	//client.hub.closeSession <- client
}

func (client *Client) read() {
	defer client.close()

	for {
		messageType, p, err := client.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[peer::user] eror: %v", err)
			}
			break
		}
		msg := &Message{messageType, p, client.Peer, client.sid}
		fmt.Println("[peer::user] Message received", msg)
		select {
		case client.hub.usersIncoming <- msg:
			fmt.Println("[peer::user] Forwarding message", msg)
		}
	}
}
