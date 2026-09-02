package main

import (
	"fmt"
	"time"
)

var brokerAddress = "localhost:8082"

func main() {

	// Agent identification
	id := generateId()

	// Ask for secret token provided with the distribution
	secretToken := readToken()

	// {D,d} keys used to shared with broker sharing with the party
	keys := generateKeys()
	sendAuthentication(brokerAddress+"/open", secretToken, &keys, id, func([]byte) {})

	fmt.Println("token", secretToken)

	// builds an interval signal used as Hearbeat
	beacon := time.NewTicker(30 * time.Second)
	defer beacon.Stop()

	for {
		select {
		case timestamp := <-beacon.C:
			handleBeaconRequest(&keys, id, timestamp)
		}
	}
}
