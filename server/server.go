package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"syscall"

	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/model"
	"golang.org/x/net/websocket"
)

const (
	// ClosingSuccess is a response message received when close endpoint is called
	ClosingSuccess = "Closing..."
)

// Version is declaration of the server protocol version that this server provides
var Version = "undefined"

/*
CliciServer is the main server class of Clici. It is the mediator between Jenkins server(s)
and the Clici clients, lowering the impact on the Jenkins server and giving
more real-time push style of notifications to the clients.
*/
type CliciServer struct {
	*http.ServeMux
	lis              net.Listener
	closedGracefully bool
	// Secret is the URL on this server that can be called to
	// gracefully shutdown the server
	Secret string
	// Port is the port which will be occupied by the server
	Port      int
	processor *Processor
}

// New creates a new Clici server behind a certain port.
// Nothing will be started until StartAndWait is called though.
func New(port int) CliciServer {
	clici := CliciServer{
		ServeMux:  http.NewServeMux(),
		Port:      port,
		processor: NewProcessorWithSupplier(jenkins.NewAPI),
	}
	return clici
}

/*
StartAndWait is the entry point after the server object has been initiated.
It will block until shutdown call is executed or until program is interrupted
*/
func (h *CliciServer) StartAndWait(started chan<- struct{}) {
	var err error
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", h.Port))
	if err != nil {
		log.Fatalf("Could not listen: %v", err)
	}
	h.lis = lis

	h.registerRandomizedShutdownHook()

	h.ServeMux.Handle("/ws", websocket.Handler(h.clientHandler))

	started <- struct{}{}
	if err = http.Serve(lis, h); err != nil && !h.closedGracefully {
		log.Fatalf("Could not start serving: %v", err)
	}
}

func (h *CliciServer) registerRandomizedShutdownHook() {
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

func (h *CliciServer) processRegistrationRequestsFromClient(id ConnectionID, newRegistrations chan<- Register, clientLeft chan<- bool, lepr *LengthEncodedProtoReaderWriter) {
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
}

func (h *CliciServer) processOutgoingUpdates(id ConnectionID, outgoingJobStateChannel <-chan model.JobState, clientLeft <-chan bool, lepr io.Writer) {
	for {
		select {
		case state := <-outgoingJobStateChannel:
			log.Printf("Publishing state %v to client behind %v", state, id)
			// TODO: just a POC, a DTO should be streamed
			_, _ = lepr.Write([]byte(fmt.Sprintf("%v", state)))
		case <-clientLeft:
			log.Printf("Connect %v left", id)
			return
		}
	}
}

func (h *CliciServer) clientHandler(ws *websocket.Conn) {
	id := ConnectionID(randomStringFromBytes(8))
	newRegistrations := make(chan Register)
	clientLeft := make(chan bool)
	outputChannel := make(chan model.JobState)
	defer close(clientLeft)
	defer close(newRegistrations)

	lepr := &LengthEncodedProtoReaderWriter{UnderlyingReadWriter: ws}

	go h.processRegistrationRequestsFromClient(id, newRegistrations, clientLeft, lepr)
	go h.processOutgoingUpdates(id, outputChannel, clientLeft, lepr)

outer:
	for {
		select {
		case requestedMappings := <-newRegistrations:
			for _, job := range requestedMappings.GetJobs() {
				h.processor.RegisterClient(id, job.ServerLocation, job.JobName, outputChannel)
			}
		case <-clientLeft:
			h.processor.UnRegisterClient(id)
			break outer
		}
	}
}

func (h *CliciServer) respondAllOk(lepr *LengthEncodedProtoReaderWriter, id ConnectionID) bool {
	response := RegisterResponse{
		Version: Version,
		Success: true,
		Connid:  id.AsString(),
	}
	if err := lepr.WriteProto(&response); err != nil {
		log.Printf("Failure responding to request: %v, terminating connection", err)
		_ = lepr.UnderlyingReadWriter.Close()
		return false
	}
	return true
}
