package main

import (
	"fmt"
	"log"

	requests "./protocol"

	"github.com/golang/protobuf/proto"
)

func DeserialiseAuth(p []byte) *requests.Auth {
	auth := &requests.Auth{}
	proto.Unmarshal(p, auth)
	return auth
}

func DeserialiseAuthAck(p []byte) *requests.AuthAck {
	auth := &requests.AuthAck{}
	proto.Unmarshal(p, auth)
	return auth
}

// GetRequest obtains the request representation
func GetRequest(data []byte) *requests.Request {
	n := len(data)
	req := &requests.Request{}
	err := proto.Unmarshal(data[:n], req)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	return req
}

func SerialiseAuth(sid int32, pk string) []byte {
	ack := &requests.Auth{
		Id:        proto.Int32(1),
		Sid:       proto.Int32(sid),
		PublicKey: proto.String(pk)}
	data, err := proto.Marshal(ack)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	return data
}

func SerialiseAuthAck(sid int32, pk string) []byte {
	ack := &requests.AuthAck{
		Id:        proto.Int32(2),
		Sid:       proto.Int32(sid),
		PublicKey: proto.String(pk)}
	data, err := proto.Marshal(ack)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	return data
}

func SerialiseAuthAck2(sid int32, port int32) []byte {
	ack := &requests.AuthAck{
		Id:        proto.Int32(4),
		Sid:       proto.Int32(sid),
		PublicKey: proto.String(""),
		Port:      proto.Int32(port),
	}
	data, err := proto.Marshal(ack)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	return data
}

// TestAuthAckSerialisation protocol asserts the authentication.
func TestAuthAckSerialisation() {
	fmt.Println("............Test_AuthAck_Serialisation..........................................")

	ack := &requests.AuthAck{
		Sid:       proto.Int32(1000),
		PublicKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaa")}

	// test ack
	data, err := proto.Marshal(ack)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	testAuth := &requests.AuthAck{}
	err = proto.Unmarshal(data, testAuth)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	// Now test and newTest contain the same data.
	if testAuth.GetSid() != ack.GetSid() {
		log.Fatalf("data mismatch %q != %q", testAuth.GetSid(), ack.GetSid())
	}
	fmt.Printf("id:%d-%d\n", testAuth.GetId(), ack.GetId())
	fmt.Println(".")
}

// TestAuthSerialisation protocol asserts the authentication.
func TestAuthSerialisation() {
	fmt.Println("............Test_Auth_Serialisation..........................................")

	auth := &requests.Auth{
		Sid:       proto.Int32(1000),
		PublicKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaa")}

	// test ack
	data, err := proto.Marshal(auth)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	testAuth := &requests.Auth{}
	err = proto.Unmarshal(data, testAuth)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	// Now test and newTest contain the same data.
	if testAuth.GetSid() != auth.GetSid() {
		log.Fatalf("data mismatch %q != %q", testAuth.GetSid(), auth.GetSid())
	}
	fmt.Println(".")
}

// protoc --go_out=. protocol/requests.proto

// TestProtocolAuthSerialisation assert the whole involved.
func TestProtocolAuthSerialisation() {
	fmt.Println("............Test_Protocol_Auth_Serialisation..........................................")
	TestAuthSerialisation()
	TestAuthAckSerialisation()
	fmt.Println(".")
}
