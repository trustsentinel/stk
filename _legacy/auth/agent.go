package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// Agent wrapper of connection associated to
// one agent connections from hosts under registration
type Agent struct {
	*Peer
	token     string
	timestamp int64
}

func (agent *Agent) close() {

	// Close the current connection
	agent.Peer.close()

	// Disconnects peer from peers pool
	agent.hub.unregister <- agent
}

func (agent *Agent) read() {
	defer agent.close()

	for {
		messageType, p, err := agent.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[peer::agent] Error: %v", err)
			}
			break
		}
		msg := &Message{messageType, p, agent.Peer, int32(0)}
		if msg.data[1] != 3 { // silent hearbeat!
			log.Println("[peer::agent] Message received from agent", msg)
		}

		select {
		case agent.hub.agentsIncoming <- msg:
			if msg.data[1] != 3 {
				log.Println("[peer::agent] Forwarding message to hub", msg)
			}
		}
	}
}
