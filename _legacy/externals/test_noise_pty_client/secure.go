package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"

	"golang.org/x/crypto/nacl/box"
)

// SecureChannel as connection wrapper
type SecureChannel struct {
	S    *[32]byte
	conn *net.TCPConn
}

func newSecureChannel(c *net.TCPConn, S [32]byte) *SecureChannel {
	return &SecureChannel{
		S:    &S,
		conn: c,
	}
}

// SecureMessage as message wrapper
type SecureMessage struct {
	msg   []byte
	nonce [24]byte
}

func (s *SecureMessage) toByteArray() []byte {
	return append(s.nonce[:], s.msg[:]...)
}

// ToSecureMessage convert byte array into struct
func ToSecureMessage(b []byte) SecureMessage {
	var nonce [24]byte
	na := b[:24]
	copy(nonce[:], na)
	message := bytes.Trim(b[24:], "\x00")
	return SecureMessage{msg: message, nonce: nonce}
}

// Write into the secure channel
func (s *SecureChannel) Write(msg []byte) (int, error) {
	var nonce [24]byte
	rand.Read(nonce[:])
	encryptedMessage := box.SealAfterPrecomputation(nil, msg, &nonce, s.S)
	sm := SecureMessage{msg: encryptedMessage, nonce: nonce}
	return s.conn.Write(sm.toByteArray())
}

// Decrypt the message
func Decrypt(encoded []byte, sharedKey [32]byte) []byte {
	fmt.Println("encoded", encoded[:])
	message, _ := base64.StdEncoding.DecodeString(string(encoded[:]))

	secureMessage := ToSecureMessage(message)
	fmt.Println("message", message[:], string(message[:]))
	fmt.Println(" -> nonce", secureMessage.nonce[:], string(secureMessage.nonce[:]))
	fmt.Println(" -> msg", secureMessage.msg[:], string(secureMessage.msg[:]))
	fmt.Println(" -> s", sharedKey, string(sharedKey[:32]))
	decryptedMessage, _ := box.OpenAfterPrecomputation(nil, secureMessage.msg, &secureMessage.nonce, &sharedKey)
	fmt.Println(" -> d", decryptedMessage, string(decryptedMessage[:]))

	encryptedMessage := box.SealAfterPrecomputation(nil, []byte("d"), &secureMessage.nonce, &sharedKey)
	fmt.Println(" -> e", encryptedMessage, string(encryptedMessage[:]))

	decryptedMessage2, _ := box.OpenAfterPrecomputation(nil, encryptedMessage, &secureMessage.nonce, &sharedKey)
	fmt.Println(" -> de", decryptedMessage2, string(decryptedMessage2[:]))
	return decryptedMessage
}

// Encrypt the messaEncryptge
func Encrypt(message []byte, sharedKey [32]byte) []byte {
	var nonce [24]byte
	rand.Read(nonce[:])
	encryptedMessage := box.SealAfterPrecomputation(nil, message, &nonce, &sharedKey)
	sm := SecureMessage{msg: encryptedMessage, nonce: nonce}
	return []byte(base64.StdEncoding.EncodeToString(sm.toByteArray()))
}

func (s *SecureChannel) Read(msg []byte) (int, error) {
	message := make([]byte, 2048)
	n, err := s.conn.Read(message)

	secureMessage := ToSecureMessage(message)
	decryptedMessage, ok := box.OpenAfterPrecomputation(nil, secureMessage.msg, &secureMessage.nonce, s.S)

	if !ok {
		return 0, errors.New("problem decrypting the message")
	}
	n = copy(msg, decryptedMessage)
	return n, err
}

/*
		const pairA = {
	            publicKey: new Uint8Array([16,44,122,53,10,77,214,164,220,19,151,64,250,13,191,5,61,6,88,138,89,1,190,18,119,176,132,39,227,233,73,35]),
	            secretKey: new Uint8Array([160,184,150,49,86,171,26,150,235,114,62,19,126,123,44,11,225,238,103,124,241,203,122,13,147,247,18,48,24,232,121,121])
	        }

	        const pairB = {
	            publicKey: new Uint8Array([149,159,254,46,1,184,215,240,224,123,198,64,123,80,0,135,113,236,213,76,117,201,253,66,12,214,30,129,219,32,11,87]),
	            secretKey: new Uint8Array([19,30,255,99,200,249,241,57,246,28,248,18,143,230,19,102,60,194,104,10,96,5,33,37,241,157,163,58,223,97,192,204])
	        }
*/
func test(A [32]byte, a [32]byte, B [32]byte, b [32]byte, S [32]byte) {
	fmt.Println("******************************************************************")
	var nonce [24]byte

	// Create a new nonce for each message sent
	rand.Read(nonce[:])
	message := []byte("hello")
	fmt.Println("Message", message, string(message))
	encryptedMessage := box.SealAfterPrecomputation(nil, message, &nonce, &S)
	fmt.Println("Encrypted", encryptedMessage, string(encryptedMessage))

	decryptedMessage, _ := box.OpenAfterPrecomputation(nil, encryptedMessage, &nonce, &S)
	fmt.Println("Decrypted", decryptedMessage, string(decryptedMessage))
	fmt.Println("******************************************************************")

	e := Encrypt([]byte("hello2"), S)
	fmt.Println("enc ->", e, string(e))

	d := Decrypt(e, S)
	fmt.Println("-> dec", d, string(d))

	e2 := Encrypt(d, S)
	fmt.Println("<- enc", e2, string(e2))

	d2 := Decrypt(e2, S)
	fmt.Println("dec <-", d2, string(d2))
}

func getRemoteKey() [32]byte {
	var public [32]byte
	A := []byte{16, 44, 122, 53, 10, 77, 214, 164, 220, 19, 151, 64, 250, 13, 191, 5, 61, 6, 88, 138, 89, 1, 190, 18, 119, 176, 132, 39, 227, 233, 73, 35}
	copy(public[:], A)
	fmt.Println("A(public)  ", A[:], string(A[:]))
	return public
}

func getPrivateRemoteKey() [32]byte {
	var private [32]byte
	a := []byte{160, 184, 150, 49, 86, 171, 26, 150, 235, 114, 62, 19, 126, 123, 44, 11, 225, 238, 103, 124, 241, 203, 122, 13, 147, 247, 18, 48, 24, 232, 121, 121}
	copy(private[:], a)
	fmt.Println("a(private)  ", a[:], string(a[:]))
	return private
}

func generatePair() ([32]byte, [32]byte) {
	var public, private [32]byte
	B := []byte{149, 159, 254, 46, 1, 184, 215, 240, 224, 123, 198, 64, 123, 80, 0, 135, 113, 236, 213, 76, 117, 201, 253, 66, 12, 214, 30, 129, 219, 32, 11, 87}
	b := []byte{19, 30, 255, 99, 200, 249, 241, 57, 246, 28, 248, 18, 143, 230, 19, 102, 60, 194, 104, 10, 96, 5, 33, 37, 241, 157, 163, 58, 223, 97, 192, 204}
	//B, b, _ := box.GenerateKey(rand.Reader)
	copy(public[:], B)
	copy(private[:], b)
	fmt.Println("B(public)  ", B[:], string(B[:]))
	fmt.Println("b(private) ", b[:], string(b[:]))
	return public, private
}

func getSharedKey(remotePublicKey [32]byte, localPrivateKey [32]byte) [32]byte {
	var sharedKey [32]byte
	box.Precompute(&sharedKey, &remotePublicKey, &localPrivateKey)
	fmt.Println("remotePublicKey ", remotePublicKey, string(remotePublicKey[:]))
	fmt.Println("localPrivateKey ", localPrivateKey, string(localPrivateKey[:]))
	fmt.Println("S(shared)  ", sharedKey, string(sharedKey[:]))
	return sharedKey
}
