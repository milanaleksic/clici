package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"crypto/rand"
	"encoding/base64"
	"net"
	"syscall"

	"golang.org/x/net/websocket"
)

type handler struct {
	*http.ServeMux
	lis    net.Listener
	secret string
	port int
}

func (h *handler) startAndWait(started chan <- struct{}) {
	var err error
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		log.Fatalf("Could not listen: %v", err)
	}

	randData := make([]byte, 48)
	rand.Read(randData)
	h.secret = base64.StdEncoding.EncodeToString(randData)
	h.ServeMux.HandleFunc(fmt.Sprintf("/%v", h.secret), h.close)

	h.ServeMux.Handle("/echo", websocket.Handler(h.echoServer))

	started <- struct{}{}
	http.Serve(lis, h)
}

func (h *handler) close(w http.ResponseWriter, r *http.Request) {
	h.lis.Close()
	log.Println("Closing the listener")
}

func (h *handler) echoServer(ws *websocket.Conn) {
	for {
		var readBytes = make([]byte, 512)
		n, err := ws.Read(readBytes)
		if err != nil {
			if err == io.EOF {
				return
			} else if strings.Contains(err.Error(), syscall.ECONNRESET.Error()) {
				ws.Close()
				return
			} else {
				log.Printf("Failure receiving: %v, terminating connection", err)
				ws.Close()
				return
			}
		}
		fmt.Printf("Received %d bytes: %v\n", n, string(readBytes))
		io.Copy(ws, strings.NewReader("Received!"))
	}
}

func main() {
	handler := &handler{ServeMux: http.NewServeMux()}
	handler.startAndWait(make(chan <- struct{}))
}
