package main

// TRequest interface of general messaging
type TRequest struct {
	Type string
	Data map[string]interface{}
}

// Message structure contains message the common scheme to
// compose messages used by Hub
type Message struct {
	messageType int
	data        []byte
	peer        *Peer
	sid         int32 `default0:0`
}
