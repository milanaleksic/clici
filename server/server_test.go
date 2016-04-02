package main

import (
	"fmt"
	"testing"

	"net/http"

	"golang.org/x/net/websocket"
)

func TestEchoServer(t *testing.T) {
	handler := &handler{ServeMux: http.NewServeMux(), port: 12345}
	started := make(chan struct{}, 0)
	go handler.startAndWait(started)
	defer func() { _, _ = http.Get(fmt.Sprintf("localhost:12345/%s", handler.secret)) }()

	<-started

	origin := "http://localhost/"
	url := "ws://localhost:12345/echo"
	ws, err := websocket.Dial(url, "ws", origin)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Dialed")
	if _, err = ws.Write([]byte("hello, world! 1")); err != nil {
		t.Fatalf("??? %v", err)
	}
	if _, err = ws.Write([]byte("hello, world! 2")); err != nil {
		t.Fatalf("??? %v", err)
	}
	if _, err = ws.Write([]byte("hello, world! 3")); err != nil {
		t.Fatalf("??? %v", err)
	}
	fmt.Println("Written")
	var msg = make([]byte, 512)
	var n int
	fmt.Println("Reading")
	if n, err = ws.Read(msg); err != nil {
		t.Fatalf("??? %v", err)
	}
	fmt.Printf("Received: %s.\n", msg[:n])
}
