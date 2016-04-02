package main

import (
	"fmt"
	"testing"

	"net/http"

	"golang.org/x/net/websocket"
	"io/ioutil"
)

func TestEchoServer(t *testing.T) {
	withRunningServer(t, func() {
		ws := dial(t)
		writeBytes(t, ws, []byte("hello, world! 1"))
		writeBytes(t, ws, []byte("hello, world! 2"))
		writeBytes(t, ws, []byte("hello, world! 3"))
		for i := 0; i < 3; i++ {
			read := readBytes(t, ws)
			expected := "Received!"
			if string(read) != expected {
				t.Fatalf("[%v] != [%v]", string(read), expected)
			}
		}
	})
}

func withRunningServer(t *testing.T, callback func()) {
	handler := &handler{ServeMux: http.NewServeMux(), port: 12345}
	started := make(chan struct{}, 0)
	go handler.startAndWait(started)
	defer func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:12345/%s", handler.secret))
		if err != nil {
			t.Fatalf("Err: %v", err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if string(b) != ClosingSuccess {
			t.Errorf("Graceful server shutdown failed: %v", string(b))
		}
	}()

	<-started

	callback()
}

func dial(t *testing.T) (ws *websocket.Conn) {
	origin := "http://ignored/"
	url := "ws://localhost:12345/echo"
	ws, err := websocket.Dial(url, "ws", origin)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func writeBytes(t *testing.T, ws *websocket.Conn, bytes []byte) {
	if _, err := ws.Write(bytes); err != nil {
		t.Fatalf("??? %v", err)
	}
}

func readBytes(t *testing.T, ws *websocket.Conn) (read []byte) {
	var n int
	var msg = make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		t.Fatalf("??? %v", err)
	}
	return msg[:n]
}