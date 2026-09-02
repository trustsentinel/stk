package main

import (
	"fmt"
	"log"

	requests "./protocol"

	"github.com/golang/protobuf/proto"
)

func DeserialiseAgentAuth(p []byte) *requests.AgentAuth {
	auth := &requests.AgentAuth{}
	proto.Unmarshal(p, auth)
	return auth
}
func SerialiseAgentAuth(token string, pk string, deviceId string) []byte {
	auth := &requests.AgentAuth{
		Token:     proto.String(token),
		PublicKey: proto.String(token),
		DeviceId:  proto.String(token),
	}
	data, err := proto.Marshal(auth)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	return data
}

func TestAgentSerialisation() {
	fmt.Println("............TestAgentSerialisation..........................................")

	auth := &requests.AgentAuth{
		Token:     proto.String("mytoken"),
		PublicKey: proto.String("aaaaaaaaaaaaaaaaaaaaaaaaaa"),
		DeviceId:  proto.String("abcdefg"),
	}

	// test ack
	data, err := proto.Marshal(auth)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	testAuth := &requests.AgentAuth{}
	err = proto.Unmarshal(data, testAuth)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

	// Now test and newTest contain the same data.
	if testAuth.GetDeviceId() != auth.GetDeviceId() {
		log.Fatalf("data mismatch %q != %q", testAuth.GetDeviceId(), auth.GetDeviceId())
	}
	fmt.Printf("id:%d-%d\n", testAuth.GetId(), auth.GetId())
	fmt.Println(".")
}
