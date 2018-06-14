package main

import (
	"log"
	"time"

	"github.com/milanaleksic/clici/cmd/main/controller"
	"github.com/milanaleksic/clici/cmd/main/view"
)

type dispatcher struct {
	feedbackChannel chan view.Command
	controller      *controller.Controller
}

func (dispatcher *dispatcher) mainLoop() {
	ticker := time.NewTicker(options.Application.Refresh.Duration)
	defer ticker.Stop()
	firstRun := make(chan bool, 1)
	firstRun <- true
	for {
		select {
		case x := <-dispatcher.feedbackChannel:
			if shouldExit := dispatcher.dispatch(x); shouldExit {
				return
			}
		case <-ticker.C:
			dispatcher.controller.RefreshAllNodeInformation()
		case <-firstRun:
			dispatcher.controller.RefreshAllNodeInformation()
		}
	}
}

func (dispatcher *dispatcher) dispatch(x view.Command) bool {
	log.Printf("Dispatcher received command: %q\n", x)
	switch x.Group {
	case view.CmdShutdownGroup:
		log.Println("Bye!")
		return true
	case view.CmdCloseGroup:
		dispatcher.controller.RemoveModals()
	case view.CmdShowHelpGroup:
		dispatcher.controller.ShowHelp()
	case view.CmdOpenCurrentJobGroup:
		dispatcher.controller.VisitCurrentJob(x.Job)
	case view.CmdOpenPreviousJobGroup:
		dispatcher.controller.VisitPreviousJob(x.Job)
	case view.CmdTestsForJobGroup:
		dispatcher.controller.ShowTests(x.Job)
	case view.CmdRunJob:
		dispatcher.controller.RunJob(x.Job)
	}
	return false
}
