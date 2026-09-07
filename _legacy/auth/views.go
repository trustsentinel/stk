package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"time"
	_ "websocket"

	"github.com/gorilla/websocket"

	"github.com/gorilla/mux"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// getQR view which generates the qr image based on passcode recieved as parameter
// GET localhost:8081/auth/qr/{passcode}
func getQR(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	passcode := params["passcode"]

	qr, err := auth.GenerateQr(passcode)
	if err != nil {
		json.NewEncoder(w).Encode("error generating qr")
		return
	}
	png.Encode(w, qr)
}

// getTotpValidation view verifies if prompt TOTP code is the expected
// GET localhost:8081/auth/validate/{passcode}/{totp}
func getTotpValidation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	passcode := params["passcode"]
	totp := params["totp"]
	valid := auth.ValidateTotp(passcode, totp)
	log.Println("[views] User with", passcode, "is trying to validate with password: ", totp, ", result: ", valid)
	json.NewEncoder(w).Encode(valid)
}

// getTotpSecrets view show the current secrets stored [admin view]
// GET localhost:8081/auth/secrets
func getTotpSecrets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("[views] Get secrets ", auth.secrets, &auth.secrets)
	json.NewEncoder(w).Encode(auth.secrets)
}

// getDevice retrieve a view with the current registered devices
// GET localhost:8081/devices
func getDevices(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices.Get())
}

/// goCreateChannel instante a new channel (dedicated) with a ws client-server connection
/// common channel is bound to this dedicated channel
func goCreateChannel(incoming *chan []byte, outgoing *chan []byte, cin *chan []byte, cout *chan []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		channelID := params["channel"]
		channels[channelID] = make(chan bool)
		log.Println("Creating new channel", channelID)

		go createDedicatedChannel(channelID, 8100, channels[channelID], incoming, outgoing, cin, cout)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(channelID)
	}
}

// GetRole extracts role from header
func GetRole(r *http.Request) string {
	role := r.Header.Get("Stk-role")
	if len(role) == 0 {
		return "user"
	}
	return role
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getUserKey() ([]byte, [32]byte) {
	var public [32]byte
	A := []byte{16, 44, 122, 53, 10, 77, 214, 164, 220, 19, 151, 64, 250, 13, 191, 5, 61, 6, 88, 138, 89, 1, 190, 18, 119, 176, 132, 39, 227, 233, 73, 35}
	copy(public[:], A)
	fmt.Println("A(public)  ", A[:], string(A[:]))
	return A, public
}

func handleAdminDevices(hub *Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		token := mux.Vars(r)["token"]
		u, _ := getUserKey()
		sid := 1

		// senging authentication message to device
		data := SerialiseAuth(1, base64.StdEncoding.EncodeToString(u))
		peer := hub.agents[token]
		message := &Message{websocket.BinaryMessage, data, peer.Peer, int32(sid)}
		hub.agentsIncoming <- message
	}
}

func handleDevices(hub *Hub) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := mux.Vars(r)["token"]
		role := GetRole(r)

		log.Printf("[auth] Incoming request with token: %s, role: %s", token, role)

		// It will determine whether or not an incoming request from a different domain
		// is allowed to connect, and if it isn’t they’ll be hit with a CORS error
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		// It upgrades this connection to a WebSocket
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// Peer representation of connection
		peer := &Peer{
			hub:  hub,
			conn: ws,
			send: make(chan []byte, 1024)}

		timestamp := time.Now().Unix()
		switch role {
		case "agent":
			agent := &Agent{peer, token, timestamp}
			go agent.read()
			hub.register <- agent
			break
		default:
			client := &Client{peer, token, int32(0), timestamp}
			go client.read()
			hub.sessions <- client
		}
	}
}
