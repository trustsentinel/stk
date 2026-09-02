package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
)

// Devices general structure to store devices
type Devices struct {
	id      string
	db      *SafeTokens
	max     int
	devices []Device
}

// Device representation
type Device struct {
	Token     string `json:"token"`
	Timestamp string `json:"update"`
}

// Get retrieve the current devices
func (d *Devices) Get() []Device {
	d.SyncUp()
	return d.devices[:d.max]
}

// GenerateToken generates a token
func GenerateToken(l int) string {
	buff := make([]byte, int(math.Round(float64(l)/2)))
	rand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

// Init add placeholder devices or saved from local
// or using a external repository
func (d *Devices) Init(db *SafeTokens) {
	d.db = db
	d.devices = make([]Device, 1000)
	if db == nil {
		d.devices = []Device{
			{Token: "dcb6888d", Timestamp: "1231231232323"},
			{Token: GenerateToken(8), Timestamp: "1231231232324"},
		}
		log.Println("[devices] Initialisation with placeholder devices!", d.devices[:d.max])
		return
	}
}

// SyncUp sync with the latest updated tokens
func (d *Devices) SyncUp() {
	log.Println("[devices] SyncUp devices...")
	idx := 0
	for token, entry := range d.db.tokens {
		timestamp := int2string(int(toTimestamp(entry.updated)))
		fmt.Printf("   [%s] %s\n", token, timestamp)
		d.devices[idx] = Device{Token: token, Timestamp: timestamp}
		idx++
	}
	d.max = idx
	log.Println("[devices] found out token devices!", d.devices[:d.max])
}
