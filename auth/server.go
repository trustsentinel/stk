package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Already object initialisation
var devices = &Devices{id: "1"}
var auth = newAuth()

var cin = make(chan []byte, 1)
var cout = make(chan []byte, 1)
var incoming = make(chan []byte, 1)
var outgoing = make(chan []byte, 1)

func main() {

	auth.LoadFromSaved()

	// Hub
	db := SafeTokens{tokens: make(map[string]Registration)}
	devices.Init(&db)
	hub := newAuthHub(&db)
	go hub.Start()

	log.Println("[main] Starting server...")
	router := mux.NewRouter()

	// Devices
	router.HandleFunc("/devices/registered", getDevices).Methods("GET")

	// Channels
	router.HandleFunc("/admin/channel/message/{channel}",
		goMessageChannel(&incoming)).Methods("GET")
	router.HandleFunc("/admin/channel/create/{channel}",
		goCreateChannel(&incoming, &outgoing, &cin, &cout)).Methods("GET")
	router.HandleFunc("/channels/{channel}/ws",
		goWsChannel(&incoming, &outgoing, &cin, &cout))

	// Devices
	router.HandleFunc("/admin/devices/{token}", handleAdminDevices(hub))
	router.HandleFunc("/devices/{token}", handleDevices(hub))

	// Auth module
	router.HandleFunc("/auth/qr/{passcode}", getQR).Methods("GET")
	router.HandleFunc("/auth/validate/{passcode}/{totp}", getTotpValidation).Methods("GET")
	router.HandleFunc("/auth/secrets", getTotpSecrets).Methods("GET")

	http.Handle("/", router)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("[main] ListenAndServe: ", err)
	}
}
