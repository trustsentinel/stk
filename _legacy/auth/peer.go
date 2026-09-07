package main

import (
	"github.com/gorilla/websocket"
)

// Peer representation, contains the connection
// with a channel to set the communication
type Peer struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (p *Peer) close() {
	p.conn.Close()
}
