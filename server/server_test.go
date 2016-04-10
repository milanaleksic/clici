package server

import (
	"fmt"
	"testing"

	"net/http"

	"io/ioutil"
	"log"
	"net"

	"golang.org/x/net/websocket"
	"time"
)

func TestRegistration(t *testing.T) {
	var mapping *Mapping
	var registration interface{}
	var connid string
	withRunningServer(t, func(clici *Clici, ws *websocket.Conn) {
		mapping = clici.mapping
		request := &Register{
			Jobs: []*Register_Job{
				&Register_Job{
					ServerLocation: "localhost:8101/jenkins/",
					JobName:        "test_job_1",
				},
			},
		}
		wire := &LengthEncodedProtoReaderWriter{UnderlyingReadWriter: ws}
		if err := wire.WriteProto(request); err != nil {
			t.Fatalf("registration failed while writing request: %v", err)
		}
		response := RegisterResponse{}
		if err := wire.ReadProto(&response); err != nil {
			t.Fatalf("registration failed with error: %v", err)
		} else if !response.Success {
			t.Fatalf("registration failed")
		}
		err := retry(func() (err error) {
			tx := clici.mapping.db.Txn(false)
			registration, err = tx.First(registrationTable, "connid", response.Connid)
			if registration != nil && err == nil {
				return
			}
			return fmt.Errorf("Registration is not set")
		})
		if err != nil {
			t.Fatalf("registration did not create new record in memdb on server side: err=%v, registration=%v", err, registration)
		}
		connid = response.Connid
	})

	err := retry(func() (err error) {
		tx := mapping.db.Txn(false)
		registration, err = tx.First(registrationTable, "connid", connid)
		if registration == nil && err == nil {
			return
		}
		return fmt.Errorf("Registration is still not removed")
	})
	if err != nil {
		t.Fatalf("closing did not remove connection from memdb: err=%v, registration=%v", err, registration)
	}
}

func withRunningServer(t *testing.T, callback func(clici *Clici, ws *websocket.Conn)) {
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
	handler := NewClici(port)
	started := make(chan struct{}, 0)
	go handler.StartAndWait(started)
	defer func() {
		// FIXME: why is this sometimes returning 404?
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

	ws := dial(t, port)
	defer func() {_ = ws.Close()} ()

	callback(&handler, ws)
}

func dial(t *testing.T, port int) (ws *websocket.Conn) {
	origin := "http://ignored/"
	url := fmt.Sprintf("ws://localhost:%d/ws", port)
	ws, err := websocket.Dial(url, "ws", origin)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func retry(closure func() error) (err error) {
	for i := 0; i < 100; i++ {
		err = closure()
		if err == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	return
}