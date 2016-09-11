package main

import (
	"fmt"
	"log"

	"github.com/milanaleksic/clici/cmd/main/controller"
	"github.com/milanaleksic/clici/cmd/main/view"
	"github.com/milanaleksic/clici/jenkins"
)

// Version holds the main version string which should be updated externally when building release
var Version = "undefined"

func getAPI() (result []controller.JenkinsAPIRoot) {
	if options.Application.Mock {
		for _, aServer := range options.Jenkins {
			result = append(result, controller.JenkinsAPIRoot{
				API:  jenkins.NewMockAPI(),
				Jobs: aServer.Jobs,
			})
		}
	}
	for _, aServer := range options.Jenkins {
		result = append(result, controller.JenkinsAPIRoot{
			API:  jenkins.NewAPI(aServer.Location),
			Jobs: aServer.Jobs,
		})
	}
	return
}

func getUI(feedbackChannel chan view.Command) (ui view.View, err error) {
	view.AvoidUnicode = options.Interface.AvoidUnicode
	switch options.Interface.Mode {
	case interfaceSimple:
		ui = view.NewConsoleInterface(feedbackChannel)
	case interfaceAdvanced:
		ui, err = view.NewCUIInterface(feedbackChannel)
	}
	if err != nil || ui == nil {
		log.Println("Failure to activate advanced interface", err)
		ui = view.NewConsoleInterface(feedbackChannel)
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
		controller: &controller.Controller{
			View: ui,
			APIs: getAPI(),
		},
	}
	dispatcher.mainLoop()
}
