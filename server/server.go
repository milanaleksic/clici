package server

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"syscall"
)

const (
	// ClosingSuccess is a response message received when close endpoint is called
	ClosingSuccess = "Closing..."
)

// Version is declaration of the server protocol version that this server provides
var Version = "undefined"

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
	Secret string
	// Port is the port which will be occupied by the server
	Port    int
	mapping *Mapping
}

// NewClici creates a new Clici server behind a certain port.
// Nothing will be started until StartAndWait is called though.
func NewClici(port int) Clici {
	clici := Clici{
		ServeMux: http.NewServeMux(), Port: port,
	}
	clici.mapping = NewMapping()
	return clici
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

	h.ServeMux.Handle("/ws", websocket.Handler(h.registerHandler))

	started <- struct{}{}
	if err = http.Serve(lis, h); err != nil && !h.closedGracefully {
		log.Fatalf("Could not start serving: %v", err)
	}
}

func (h *Clici) registerRandomizedShutdownHook() {
	h.Secret = randomStringFromBytes(48)
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
	id := randomStringFromBytes(8)
	lepr := &LengthEncodedProtoReaderWriter{UnderlyingReadWriter: ws}
	newRegistrations := make(chan Register)
	clientLeft := make(chan bool)
	go func() {
		defer func() { clientLeft <- true }()
		for {
			register := Register{}
			err := lepr.ReadProto(&register)
			if err != nil {
				if err.Error() == io.EOF.Error() {
					//	 ignore
				} else if strings.Contains(err.Error(), syscall.ECONNRESET.Error()) {
					//	 ignore
				} else if strings.Contains(err.Error(), "closed network connection") {
					//	 ignore
				} else {
					log.Printf("Failure receiving: %v, terminating connection", err)
				}
				_ = lepr.UnderlyingReadWriter.Close()
				return
			}

			newRegistrations <- register

			if !h.respondAllOk(lepr, id) {
				return
			}
		}
	}()
	for {
		select {
		case requestedMappings := <-newRegistrations:
			h.mapping.RegisterClient(id, requestedMappings)
		case <-clientLeft:
			h.mapping.UnRegisterClient(id)
			break
		}
	}
}

func (h *Clici) respondAllOk(lepr *LengthEncodedProtoReaderWriter, id string) bool {
	response := RegisterResponse{
		Version: Version,
		Success: true,
		Connid:  id,
	}
	if err := lepr.WriteProto(&response); err != nil {
		log.Printf("Failure responding to request: %v, terminating connection", err)
		_ = lepr.UnderlyingReadWriter.Close()
		return false
	}
	return true
}
