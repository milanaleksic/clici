package main

import (
	"fmt"
	"log"

	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/view"
	"github.com/milanaleksic/clici/server"
)

// Version holds the main version string which should be updated externally when building release
var Version = "undefined"

func init() {
	server.Version = Version
}

func getAPI() jenkins.API {
	if options.Application.Mock {
		return jenkins.NewMockAPI()
	}
	return jenkins.NewAPI(options.Jenkins.Location)
}

func getUI(feedbackChannel chan view.Command) (ui view.View, err error) {
	view.AvoidUnicode = options.Interface.AvoidUnicode
	switch options.Interface.Mode {
	case interfaceSimple:
		ui, err = view.NewConsoleInterface(feedbackChannel)
	case interfaceAdvanced:
		ui, err = view.NewCUIInterface(feedbackChannel)
	}
	if err != nil || ui == nil {
		log.Println("Failure to activate advanced interface", err)
		ui, err = view.NewConsoleInterface(feedbackChannel)
	}
	return
}

func main() {
	if *options.CommandLine.showVersion {
		fmt.Printf("clici version: %v\n", Version)
		return
	}
	setupLog()
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()
	var feedbackChannel = make(chan view.Command)
	ui, err := getUI(feedbackChannel)
	if err != nil {
		log.Fatal("Failure to boot interface", err)
	}
	dispatcher := &dispatcher{
		feedbackChannel: feedbackChannel,
		controller: &controller{
			View:      ui,
			API:       getAPI(),
			KnownJobs: options.Jenkins.Jobs,
		},
	}
	dispatcher.mainLoop()
}
