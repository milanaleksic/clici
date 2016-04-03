package server

import (
	"fmt"
	"testing"

	"net/http"

	"io/ioutil"
	"log"
	"net"

	"golang.org/x/net/websocket"
	"github.com/golang/protobuf/proto"
	"encoding/binary"
	"bytes"
)

func TestEchoServer(t *testing.T) {
	withRunningServer(t, func(ws *websocket.Conn) {

		request := &Register{
			Jobs: []*Register_Job{
				&Register_Job{
					ServerLocation: "localhost:8101/jenkins/",
					JobName: "test_job_1",
				},
			},
		}
		marshalled, err := proto.Marshal(request)
		if err != nil {
			t.Fatalf("Could not marshal message: %v", err)
		}
		writeMessage(t, ws, marshalled)
		read := readBytes(t, ws)
		expected := "Received!"
		if string(read) != expected {
			t.Fatalf("[%v] != [%v]", string(read), expected)
		}
	})
}

func withRunningServer(t *testing.T, callback func(ws *websocket.Conn)) {
	port := 8080
	for ; port <= 8100; port++ {
		lis, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
		if err != nil {
			log.Printf("Skipping port %d since it can't be used: %v", port, err)
		} else {
			_ = lis.Close()
			break
		}
	}
	if port == 8101 {
		t.Fatalf("Could not execute test since all testing ports are occupied or forbidden (8080...8100)")
	}
	log.Printf("Using port %d", port)
	handler := &Clici{ServeMux: http.NewServeMux(), Port: port}
	started := make(chan struct{}, 0)
	go handler.StartAndWait(started)
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

	callback(dial(t, port))
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

func writeMessage(t *testing.T, ws *websocket.Conn, data []byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, int32(len(data)))
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	if _, err := ws.Write(buf.Bytes()); err != nil {
		t.Fatalf("write bytes for length encoding failed: %v", err)
	}

	if _, err := ws.Write(data); err != nil {
		t.Fatalf("write bytes for data failed: %v", err)
	}
}

func readBytes(t *testing.T, ws *websocket.Conn) (read []byte) {
	var n int
	var msg = make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		t.Fatalf("read bytes failed: %v", err)
	}
	return msg[:n]
}
