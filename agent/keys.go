package main

import "fmt"

// GetDeviceKey returns the placeholder key for the device
func GetDeviceKey() ([]byte, [32]byte) {
	var public [32]byte
	B := []byte{149, 159, 254, 46, 1, 184, 215, 240, 224, 123, 198, 64, 123, 80, 0, 135, 113, 236, 213, 76, 117, 201, 253, 66, 12, 214, 30, 129, 219, 32, 11, 87}
	copy(public[:], B)
	fmt.Println("B(public)  ", B[:], string(B[:]))
	return B, public
}

// Keys is the pair of 32 bytes
type Keys struct {
	public  [32]byte
	private [32]byte
}

func (keys *Keys) generate() {
	keys.public = to32([]byte{149, 159, 254, 46, 1, 184, 215, 240, 224, 123, 198, 64, 123, 80, 0, 135, 113, 236, 213, 76, 117, 201, 253, 66, 12, 214, 30, 129, 219, 32, 11, 87})
	keys.private = to32([]byte{19, 30, 255, 99, 200, 249, 241, 57, 246, 28, 248, 18, 143, 230, 19, 102, 60, 194, 104, 10, 96, 5, 33, 37, 241, 157, 163, 58, 223, 97, 192, 204})
}
