package server

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
	// ClosingSuccess is a response message received when close endpoint is called
	ClosingSuccess = "Closing..."
)

/*
Clici is the main server class of Clici. It is the mediator between Jenkins server(s)
and the Clici clients, lowering the impact on the Jenkins server and giving
more real-time push style of notifications to the clients.
 */
type Clici struct {
	*http.ServeMux
	lis              net.Listener
	closedGracefully bool
	// Secret is the URL on this server that can be called to
	// gracefully shutdown the server
	Secret           string
	// Port is the port which will be occupied by the server
	Port             int
}

/*
StartAndWait is the entry point after the server object has been initiated.
It will block until shutdown call is executed or until program is interrupted
 */
func (h *Clici) StartAndWait(started chan<- struct{}) {
	var err error
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", h.Port))
	if err != nil {
		log.Fatalf("Could not listen: %v", err)
	}
	h.lis = lis

	h.registerRandomizedShutdownHook()

	h.ServeMux.Handle("/register", websocket.Handler(h.registerHandler))

	started <- struct{}{}
	if err = http.Serve(lis, h); err != nil && !h.closedGracefully {
		log.Fatalf("Could not start serving: %v", err)
	}
}

func (h *Clici) registerRandomizedShutdownHook() {
	randData := make([]byte, 48)
	_, err := rand.Read(randData)
	if err != nil {
		log.Fatalf("Could not generate random secret: %v", err)
	}
	h.Secret = base64.StdEncoding.EncodeToString(randData)
	h.ServeMux.HandleFunc(fmt.Sprintf("/%v", h.Secret), func(w http.ResponseWriter, r *http.Request) {
		h.closedGracefully = true
		log.Println("Closing the listener")
		if _, err := w.Write([]byte(ClosingSuccess)); err != nil {
			log.Printf("Not able to send ClosingSuccess message to the client, %v", err)
		}
		if err := h.lis.Close(); err != nil {
			log.Fatalf("Not able to shutdown server gracefully, %v", err)
		}
	})
}

func (h *Clici) registerHandler(ws *websocket.Conn) {
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
