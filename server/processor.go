package server

import (
	"log"

	"github.com/milanaleksic/clici/controller"
	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/model"
	"github.com/milanaleksic/clici/view"
)

// APISupplier is a supplier of an API, in case one doesn't want to use default implementations
type APISupplier func(serverLocation string) jenkins.API

// Processor is able to wrap transparently controllers and get which mappings need to be updated with which states
type Processor struct {
	mapping     *Mapping
	controllers map[string](*controller.Controller)
	apiSupplier APISupplier
	listeners   map[ConnectionID](chan<- model.JobState)
}

// NewProcessorWithSupplier is able to create Processor with the custom supplier
func NewProcessorWithSupplier(apiSupplier APISupplier) *Processor {
	return &Processor{
		apiSupplier: apiSupplier,
		mapping:     NewMapping(),
		controllers: make(map[string](*controller.Controller)),
		listeners:   make(map[ConnectionID](chan<- model.JobState)),
	}
}

// ProcessMappings is the main call for processor which executes a blocking call on all controllers and updates
// which connections need to be updated with which states
func (processor *Processor) ProcessMappings() {
	registrationsPerServer := processor.mapping.GetAllUniqueJobs()
	log.Printf("Known Registrations: %v", registrationsPerServer)
	for server, registrations := range registrationsPerServer {
		cont, ok := processor.controllers[server]
		if !ok {
			cont = &controller.Controller{
				APIs: []controller.JenkinsAPIRoot{
					{
						API:    processor.apiSupplier(server),
						Server: server,
						Jobs:   registrations,
					},
				},
				View: view.CallbackAsView(processor.processState(server)),
			}
			processor.controllers[server] = cont
		}
		cont.RefreshNodeInformation(registrations)
	}
	return
}

func (processor *Processor) processState(server string) func(state *model.State) {
	return func(state *model.State) {
		resp := make(map[ConnectionID][]model.JobState)
		log.Printf("State received: %v", state)
		for _, jobState := range state.JobStates {
			connectionIds := processor.mapping.FindAllRegisteredConnectionsForServerAndJob(server, jobState.JobName)
			for _, connectionID := range connectionIds {
				knownJobStatesPerConnID, ok := resp[connectionID]
				if !ok {
					knownJobStatesPerConnID = make([]model.JobState, 0)
				}
				resp[connectionID] = append(knownJobStatesPerConnID, jobState)
			}
		}
		for id, models := range resp {
			listener, ok := processor.listeners[id]
			if !ok {
				log.Printf("No listener found for id: %v; all known listeners: %v", id, processor.listeners)
				continue
			}
			for _, m := range models {
				listener <- m
			}
		}
	}
}

// RegisterClient will register client in the in-memory database and will register the channel as recipient of state changes
func (processor *Processor) RegisterClient(id ConnectionID, serverLocation string, jobName string, outputChannel chan<- model.JobState) {
	processor.mapping.RegisterClient(id, registration{
		ConnectionID:   id,
		ServerLocation: serverLocation,
		JobName:        jobName,
	})
	processor.listeners[id] = outputChannel
}

// UnRegisterClient will remove all mappings
func (processor *Processor) UnRegisterClient(id ConnectionID) {
	processor.mapping.UnRegisterClient(id)
	if mapping, ok := processor.listeners[id]; ok {
		close(mapping)
	}
	processor.listeners[id] = nil
}
