package main

import (
	"log"
	"time"
	"websocket"

	requests "./protocol"
	"github.com/flynn/noise"
	"github.com/golang/protobuf/proto"
)

var sessions = make(map[int]bool)

// sendBeacon ... sends the authentication request to broker
func sendBeacon(address string, keys *noise.DHKey, deviceID string) {
	sendAuthentication(address, "", keys, deviceID, func(payload []byte) {
		ack := &requests.AuthAck{}
		proto.Unmarshal(payload, ack)

		// already session opened
		if _, ok := sessions[int(ack.GetPort())]; ok {
			return
		}

		// create channel
		createClientPlainChannel(ack.GetAddress(), int(ack.GetPort()), "/ws", func(in []byte, size int) []byte { return in },
			func(conn *websocket.Conn) {
				pttyAttach(conn)
			})
	})
}

func handleBeaconRequest(keys *noise.DHKey, deviceID string, timestamp time.Time) {
	log.Println("[auth] Sending beacon, timestamp:", timestamp)
	go sendBeacon(brokerAddress+"/open", keys, deviceID)
}
