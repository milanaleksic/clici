package server

import (
	"log"

	"github.com/milanaleksic/clici/controller"
	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/model"
	"github.com/milanaleksic/clici/view"
)

// APISupplier is a supplier of an API, in case one doesn't want to use default implementations
type APISupplier func() jenkins.API

// Processor is able to wrap transparently controllers and get which mappings need to be updated with which states
type Processor struct {
	mapping     *Mapping
	controllers map[string](*controller.Controller)
	apiSupplier APISupplier
}

// NewProcessorWithSupplier is able to create Processor with the custom supplier
func NewProcessorWithSupplier(apiSupplier APISupplier) (processor *Processor) {
	return &Processor{
		apiSupplier: apiSupplier,
		mapping: NewMapping(),
		controllers: make(map[string](*controller.Controller)),
	}
}

// ProcessMappings is the main call for processor which executes a blocking call on all controllers and updates
// which connections need to be updated with which states
func (processor *Processor) ProcessMappings() (resp map[ConnectionID][]model.JobState) {
	registrationsPerServer := processor.mapping.GetAllUniqueJobs()
	log.Printf("Known Registrations: %v", registrationsPerServer)
	resp = make(map[ConnectionID][]model.JobState)
	for server, registrations := range registrationsPerServer {
		cont, ok := processor.controllers[server]
		if !ok {
			cont = &controller.Controller{
				API: processor.apiSupplier(),
				View: view.CallbackAsView(processor.processState(server, resp)),
			}
			processor.controllers[server] = cont
		}
		cont.RefreshNodeInformation(registrations)
	}
	return
}

func (processor *Processor) processState(server string, resp map[ConnectionID][]model.JobState) (func(state *model.State)) {
	return func(state *model.State) {
		for _, jobState := range state.JobStates {
			connectionIds := processor.mapping.FindAllRegisteredConnectionsForServerAndJob(server, jobState.JobName)
			log.Printf("Connections that should receive job state updates: %v", connectionIds)
			for _, connectionID := range connectionIds {
				knownJobStatesPerConnID, ok := resp[connectionID]
				if !ok {
					knownJobStatesPerConnID = make([]model.JobState, 0)
				}
				resp[connectionID] = append(knownJobStatesPerConnID, jobState)
			}
		}
	}
}