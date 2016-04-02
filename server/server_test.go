package server

import (
	"fmt"
	"testing"

	"net/http"

	"io/ioutil"
	"log"
	"net"

	"golang.org/x/net/websocket"
)

func TestEchoServer(t *testing.T) {
	withRunningServer(t, func(port int) {
		ws := dial(t, port)
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

func withRunningServer(t *testing.T, callback func(port int)) {
	port := 8080
	for ; port < 8100; port++ {
		lis, err := net.ListenTCP(fmt.Sprintf("localhost:%d", port), nil)
		if err != nil {
			log.Printf("Skipping port %d since it can't be used", port)
			_ = lis.Close()
			break
		}
	}
	handler := &Clici{ServeMux: http.NewServeMux(), Port: port}
	started := make(chan struct{}, 0)
	go handler.startAndWait(started)
	defer func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/%s", port, handler.Secret))
		if err != nil {
			t.Fatalf("Err: %v", err)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Graceful server shutdown failed: %v", err)
		} else if string(b) != ClosingSuccess {
			t.Errorf("Graceful server shutdown failed: %v", string(b))
		}
	}()

	<-started

	callback(port)
}

func dial(t *testing.T, port int) (ws *websocket.Conn) {
	origin := "http://ignored/"
	url := fmt.Sprintf("ws://localhost:%d/register", port)
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
