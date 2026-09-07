package main

import (
	"encoding/binary"
	"time"
)

// EncodeTimestamp serialise the timestamp
func EncodeTimestamp(t time.Time) []byte {
	buf := make([]byte, 8)
	u := uint64(t.Unix())
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
