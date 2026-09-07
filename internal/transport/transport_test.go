package transport

import (
	"testing"
	"time"
)

func TestPipeRoundTrip(t *testing.T) {
	a, b := Pipe()
	if err := a.WriteMsg([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	got, err := b.ReadMsg()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q, want hello", got)
	}
	if err := b.WriteMsg([]byte("world")); err != nil {
		t.Fatal(err)
	}
	got, err = a.ReadMsg()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "world" {
		t.Fatalf("got %q, want world", got)
	}
}

func TestPipeIsolatesBuffer(t *testing.T) {
	a, b := Pipe()
	msg := []byte("mutable")
	if err := a.WriteMsg(msg); err != nil {
		t.Fatal(err)
	}
	msg[0] = 'X' // mutate after send; receiver must be unaffected
	got, err := b.ReadMsg()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "mutable" {
		t.Fatalf("got %q, want mutable (buffer not isolated)", got)
	}
}

func TestPipeCloseUnblocksRead(t *testing.T) {
	a, _ := Pipe()
	go func() {
		time.Sleep(10 * time.Millisecond)
		a.Close()
	}()
	if _, err := a.ReadMsg(); err == nil {
		t.Fatal("expected error after close")
	}
}
