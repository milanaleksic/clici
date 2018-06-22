package server

import (
	"fmt"
	"testing"

	"net/http"

	"io/ioutil"
	"log"
	"net"

	"io"
	"time"

	"github.com/milanaleksic/clici/jenkins"
	"golang.org/x/net/websocket"
)

func TestRegistration(t *testing.T) {
	var mapping *Mapping
	var connID string
	withRunningServer(t, func(clici *CliciServer, ws *websocket.Conn) {
		mapping = clici.processor.mapping
		request := &Register{
			Jobs: []*Register_Job{
				{
					ServerLocation: "localhost:8101/jenkins/",
					JobName:        "job1",
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
			t.Fatal("registration failed")
		}
		if err := assertConnectionRegisteredInMapping(clici.processor.mapping, response.Connid, true); err != nil {
			t.Fatalf("registration did not create new record in memdb on server side: err=%v", err)
		}

		clici.processor.ProcessMappings()

		readStateFromWire(t, ws)

		connID = response.Connid
	})

	if err := assertConnectionRegisteredInMapping(mapping, connID, false); err != nil {
		t.Fatalf("closing did not remove connection from memdb: err=%v", err)
	}
}

func readStateFromWire(t *testing.T, ws *websocket.Conn) {
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
	doneChan := make(chan bool)
	go func() {
		msg := make([]byte, 1024)
		for {
			n, err := ws.Read(msg)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					t.Fatalf("Error while reading response from server: %v", err)
				}
			}
			log.Printf("[WIRE] %v", string(msg[:n]))
			if n < 1024 {
				break
			}
		}
		doneChan <- true
	}()

	for {
		select {
		case <-timer.C:
			t.Fatal("Failed because timer expired!")
		case <-doneChan:
			fmt.Println("Done")
			return
		}
	}
}

func assertConnectionRegisteredInMapping(mapping *Mapping, connID string, shouldExist bool) error {
	return retry(func() (err error) {
		tx := mapping.db.Txn(false)
		registration, err := tx.First(registrationTable, "connid", connID)
		if shouldExist && registration == nil {
			return fmt.Errorf("Registration is not set: %v", registration)
		} else if (!shouldExist) && registration != nil {
			return fmt.Errorf("Registration is set: %v", registration)
		} else if err != nil {
			return fmt.Errorf("Error encountered: %v", err)
		}
		return
	})
}

func withRunningServer(t *testing.T, callback func(clici *CliciServer, ws *websocket.Conn)) {
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
		t.Fatal("Could not execute test since all testing ports are occupied or forbidden (8080...8100)")
	}
	log.Printf("Using port %d", port)

	handler := New(port)
	api := testAPI{color: "blue"}
	handler.processor.apiSupplier = func(serverLocation string, username, server string) jenkins.API {
		return &api
	}

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
	defer func() {
		_ = ws.Close()
	}()

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
