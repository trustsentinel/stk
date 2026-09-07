package main

import (
	"crypto/rand"
	"encoding/base64"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

//
// Cryptographic and randomness functions
// reference: https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/

// DecodeKey Decode a base64 string and returns the 32 bytes which represents the key
func DecodeKey(pk string) [32]byte {
	var k [32]byte
	key, _ := base64.StdEncoding.DecodeString(pk)
	copy(k[:], key[:4])
	return k
}

// GenerateRandomBytes returns securely generated random bytes.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func to32(buffer []byte) [32]byte {
	var public [32]byte
	copy(public[:], buffer)
	return public
}

//
// Other auxiliar functions
//

func toTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func int2string(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

//
// Network functionality
//
func checkPortIsBusy(address string, port int) bool {
	conn, _ := net.DialTimeout("tcp", net.JoinHostPort(address, int2string(port)), 1000)
	if conn != nil {
		conn.Close()
		return false
	}
	return true
}

// CloseIfInterrupt handles the so signal and handle the
// system interruption over application
func CloseIfInterrupt() chan os.Signal {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	return interrupt
}
