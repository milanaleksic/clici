package main

import (
	"fmt"
	"time"
	"flag"
	"strings"
)

type View interface {
	PresentState(state *State)
	Close()
}

type Options struct {
	Jobs            []string
	Server          string
	SimpleInterface bool
	Mock            bool
}

var options Options

func init() {
	jobs := flag.String("jobs", "", "CSV of all jobs on the server you want to track")
	server := flag.String("server", "http://jenkins", "URL of the Jenkins server")
	simpleInterface := flag.Bool("simple", false, "Force simple interface (keeps feeding into console)")
	mock := flag.Bool("mock", false, "Use mocked data to see how program behaves")
	flag.Parse()
	options = Options{
		Jobs: strings.Split(*jobs, ","),
		Server: *server,
		SimpleInterface: *simpleInterface,
		Mock: *mock,
	}
}

func mainLoop(shutdownChannel chan bool, ui *View) {
	api := JenkinsApi{
		ServerLocation:options.Server,
	}
	controller := Controller{
		View:*ui,
		API: api,
		KnownJobs: options.Jobs,
	}
	ticker := time.NewTicker(15 * time.Second)
	firstRun := make(chan bool, 1)
	firstRun <- true
	for {
		select {
		case <-shutdownChannel:
			fmt.Println("Bye!")
			ticker.Stop()
			return
		case <-ticker.C:
			controller.RefreshNodeInformation()
		case <-firstRun:
			controller.RefreshNodeInformation()
		}
	}
}

func main() {
	var shutdownChannel = make(chan bool)
	var ui View
	var err error
	if options.SimpleInterface {
		ui, err = NewConsoleInterface(shutdownChannel)
	} else {
		ui, err = NewCUIInterface(shutdownChannel)
		if err != nil {
			fmt.Println("Failure to advanced interface", err)
			ui, err = NewConsoleInterface(shutdownChannel)
		}
	}
	if err != nil {
		fmt.Println("Failure to boot interface", err)
	}
	if options.Mock {
		mockLoop(shutdownChannel, &ui)
	} else {
		mainLoop(shutdownChannel, &ui)
	}
}

func mockLoop(shutdownChannel chan bool, ui *View) {
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				mockData(ui)
			}
		}
	}()
	<-shutdownChannel
}

func mockData(ui *View) {
	(*ui).PresentState(&State{
		JobStates: []JobState{
			JobState{
				PreviousState: Success,
				JobName: "test_job",
				Culprits: []string{"milanale", "unknown1" },
				Time: "2 minutes more",
				Building: true,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "123456789012345678",
				Culprits: []string{"test" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "before_processing",
				Culprits: []string{"milanale", "unknown1" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "123456789012345678",
				Culprits: []string{"testing_team2", "unknown1" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Success,
				JobName: "reactive_make_sure",
				Culprits: []string{"milanale", "unknown1" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "cool_background",
				Culprits: []string{"milanale", "unknown1" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "random_name",
				Culprits: []string{"unknownX", "unknown1" },
				Time: "12 mins more than expected",
				Building: true,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Success,
				JobName: "test8",
				Culprits: []string{"milanale" },
				Time: "12 mins more than expected",
				Building: true,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "clear",
				Culprits: []string{"milanale", "unknown1" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Failure,
				JobName: "testing_mock",
				Culprits: []string{"milanale", "unknown1" },
				Time: "12 mins more than expected",
				Building: false,
				CausesFriendly: string("unknown2, unknown3"),
			},
			JobState{
				PreviousState: Success,
				JobName: "testing_mock2",
				Culprits: []string{"unknown1" },
				Time: "2 mins left",
				Building: false,
				CausesFriendly: string("unknown3"),
			},
		},
	})
}