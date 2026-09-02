package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync/atomic"

	requests "./protocol"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

var _sid int32

// Hub handles registration acts like a
// peer listener.
type Hub struct {
	agents         map[string]*Agent
	sids           map[int32]*Session
	agentsIncoming chan *Message
	usersIncoming  chan *Message
	register       chan *Agent
	unregister     chan *Agent
	sessions       chan *Client
	tokens         *SafeTokens
}

// Session structure contains information user-device session
type Session struct {
	token     string
	secretKey string
	publicKey string
	sid       int32
	peer      *Client
}

func newAuthHub(tokens *SafeTokens) *Hub {
	return &Hub{
		agents:         make(map[string]*Agent),
		sids:           make(map[int32]*Session),
		agentsIncoming: make(chan *Message, 1000),
		usersIncoming:  make(chan *Message, 1000),
		register:       make(chan *Agent, 100),
		unregister:     make(chan *Agent, 100),
		sessions:       make(chan *Client, 100),
		tokens:         tokens}
}

func handleAEMessage(hub *Hub, message *Message) {
	fmt.Println("[hub::ae] Handle message", message.data)
	if message.messageType != websocket.BinaryMessage {
		fmt.Println("[hub::ae] not a binary message")
		return
	}

	req := GetRequest(message.data)
	fmt.Println("[auth] Received command", req.GetId())
	switch req.GetId() {
	case 1: // auth is created by server, it needs to be replicated to agent
		message.peer.conn.WriteMessage(websocket.BinaryMessage, message.data)
	case 2:
		authAck := DeserialiseAuthAck(message.data)
		fmt.Printf("[auth] received command 'authAck' from sid= %d, key=%s\n", authAck.GetSid(), authAck.GetPublicKey())

		session := hub.sids[authAck.GetSid()]

		log.Println("[auth] Got the session:", session)
		channels[session.token] = make(chan bool)

		log.Println("[auth] Creating new channel", session.token)
		rand.Seed(42)
		port := 8100
		if checkPortIsBusy("localhost", port) {
			port += rand.Intn(10)
		}

		go createDedicatedChannel(session.token, port, channels[session.token], &incoming, &outgoing, &cin, &cout)

		log.Println("[auth] Sending channel indications to agent", session.token, port)
		authAck2Bytes := SerialiseAuthAck2(authAck.GetSid(), int32(port))
		hub.agentsIncoming <- &Message{websocket.BinaryMessage, authAck2Bytes, hub.agents[session.token].Peer, message.sid}
	case 4:
		authAck := DeserialiseAuthAck(message.data)
		fmt.Printf("[auth] received command 'authAck' from sid= %d, key=%s, port=%d\n", authAck.GetSid(), authAck.GetPublicKey(), authAck.GetPort())
		message.peer.conn.WriteMessage(websocket.BinaryMessage, message.data)
	default:
		fmt.Println("[auth] unknown command", message)
		//message.peer.conn.WriteMessage(websocket.BinaryMessage, message.data)
	}
}

func (hub *Hub) Start() {
	for {
		select {
		// Agent registration
		case agent := <-hub.register:
			if _, ok := hub.agents[agent.token]; ok {
				log.Println("[hub] Agent already registered") //@todo close?
			} else {
				hub.agents[agent.token] = agent
				log.Println("[hub] Register new agent:", agent.token)
				go hub.tokens.Update(agent.token)
			}
		// Agent releasing
		case agent := <-hub.unregister:
			if _, ok := hub.agents[agent.token]; ok {
				delete(hub.agents, agent.token)
				log.Printf("[hub::agent]  Agent %s unregistered\n", agent.token)
				go hub.tokens.Remove(agent.token)
			}
		// AE
		case msg := <-hub.agentsIncoming:
			if msg.messageType == websocket.BinaryMessage {
				handleAEMessage(hub, msg)
			} else {
				log.Println("[hub::agent]  Message is not binary")
			}

		// Users registration
		case client := <-hub.sessions:
			handleSession(hub, client)

		// UE
		case msg := <-hub.usersIncoming:
			if msg.messageType == websocket.TextMessage {
				handleUEMessage(hub, msg)
			} else {
				log.Println("[hub::users]  Message is not text based")
			}
		}
	}
}

func handleSession(hub *Hub, client *Client) {
	ack := &TRequest{Type: "auth"}

	// check if device token is still available
	if _, ok := hub.agents[client.token]; ok {
		sid := atomic.AddInt32(&_sid, 1)
		log.Printf("[hub] Session created, sid: %d\n", sid)
		client.sid = sid // @todo: it should be mapped in case client is connected to multiple devices (same with tokens)
		hub.sids[sid] = &Session{token: client.token, sid: sid, peer: client}

		// Client session has been created
		ack.Data = map[string]interface{}{
			"sid": strconv.FormatInt(int64(sid), 10), // int32->string
		}
	} else {
		log.Println("[hub] Agent was not found or registered yet")
		ack.Data = map[string]interface{}{
			"message": "agent was not found",
		}
	}
	data, _ := json.Marshal(ack)
	client.Peer.conn.WriteMessage(websocket.TextMessage, data)
}

func handleUEMessage(hub *Hub, message *Message) {
	ack := &TRequest{}
	fmt.Println("[hub::ue] Handle message", message.data)
	if message.messageType == websocket.BinaryMessage {
		fmt.Println("[hub::ue] not a text message")
		return
	}

	var req TRequest
	json.Unmarshal([]byte(string(message.data)), &req)
	switch req.Type {
	case "auth":

		// Client should not send authentication if session was not longer stablished
		_sid := message.sid
		if _sid == 0 {
			ack.Type = "auth"
			ack.Data = map[string]interface{}{
				"message": "client lost the session information",
			}
			data, _ := json.Marshal(ack)
			message.peer.conn.WriteMessage(websocket.TextMessage, data)
			return
		}

		// Send authentication data to the system's agent
		// with static public key from client stored on session.
		// Extra authentication also is sent in order to establish access control feature.
		session, _ := hub.sids[_sid]
		if agent, ok := hub.agents[session.token]; ok {
			publicKey := req.Data["key"].(string)
			authData := req.Data["authdata"].(string)
			log.Println("[hub::user] Authentication with authdata, key: ", publicKey, ", authdata:", authData)

			auth := &requests.Auth{
				Id:        proto.Int32(1),
				Sid:       proto.Int32(_sid),
				PublicKey: proto.String(publicKey),
			}

			data, err := proto.Marshal(auth)
			if err != nil {
				log.Fatal("[hub] marshaling error: ", err)
			}

			log.Printf("[hub] Sending packet %s\n", data)
			err = agent.conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}
