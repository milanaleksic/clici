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

const (
	ClosingSuccess = "Closing..."
)

type handler struct {
	*http.ServeMux
	lis              net.Listener
	secret           string
	port             int
	closedGracefully bool
}

func (h *handler) startAndWait(started chan<- struct{}) {
	var err error
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", h.port))
	if err != nil {
		log.Fatalf("Could not listen: %v", err)
	}
	h.lis = lis

	randData := make([]byte, 48)
	_, err = rand.Read(randData)
	if err != nil {
		log.Fatalf("Could not generate random secret: %v", err)
	}
	h.secret = base64.StdEncoding.EncodeToString(randData)
	h.ServeMux.HandleFunc(fmt.Sprintf("/%v", h.secret), h.close)

	h.ServeMux.Handle("/echo", websocket.Handler(h.echoHandler))

	started <- struct{}{}
	if err = http.Serve(lis, h); err != nil && !h.closedGracefully {
		log.Fatalf("Could not start serving: %v", err)
	}
}

func (h *handler) close(w http.ResponseWriter, r *http.Request) {
	h.closedGracefully = true
	log.Println("Closing the listener")
	w.Write([]byte(ClosingSuccess))
	if err := h.lis.Close(); err != nil {
		log.Fatalf("Not able to shutdown server gracefully, %v", err)
	}
}

func (h *handler) echoHandler(ws *websocket.Conn) {
	var readBytes = make([]byte, 512)
	for {
		n, err := ws.Read(readBytes)
		if err != nil {
			if err == io.EOF {
				return
			} else if strings.Contains(err.Error(), syscall.ECONNRESET.Error()) {
				_ = ws.Close()
				return
			} else {
				log.Printf("Failure receiving: %v, terminating connection", err)
				_ = ws.Close()
				return
			}
		}
		fmt.Printf("Received %d bytes: %v\n", n, string(readBytes))
		_, err = io.Copy(ws, strings.NewReader("Received!"))
		if err != nil {
			log.Printf("Failure echoing back: %v, terminating connection", err)
			_ = ws.Close()
			return
		}
	}
}

func main() {
	handler := &handler{ServeMux: http.NewServeMux()}
	handler.startAndWait(make(chan<- struct{}))
}
