package server

import (
	"log"

	"github.com/milanaleksic/clici/controller"
	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/model"
	"github.com/milanaleksic/clici/view"
)

type APISupplier func() jenkins.API

type Processor struct {
	mapping     *Mapping
	controllers map[string](*controller.Controller)
	apiSupplier APISupplier
}

func NewProcessorWithSupplier(apiSupplier APISupplier) (processor *Processor) {
	return &Processor{
		apiSupplier: apiSupplier,
		mapping: NewMapping(),
		controllers: make(map[string](*controller.Controller)),
	}
}

func (processor *Processor) ProcessMappings() (resp map[ConnectionID][]model.JobState) {
	registrationsPerServer := processor.mapping.GetAllUniqueJobs()
	log.Printf("Known Registrations: %v", registrationsPerServer)
	resp = make(map[ConnectionID][]model.JobState)
	for server, registrations := range registrationsPerServer {
		var knownJobs []string
		for _, reg := range registrations {
			knownJobs = append(knownJobs, reg.JobName)
		}
		cont, ok := processor.controllers[server]
		if !ok {
			cont = &controller.Controller{
				API: processor.apiSupplier(),
				View: view.CallbackAsView(processor.processState(server, resp)),
			}
			processor.controllers[server] = cont
		}
		cont.RefreshNodeInformation(knownJobs)
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