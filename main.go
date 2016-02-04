package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type View interface {
	PresentState(state *State)
	Close()
}

type Options struct {
	Jobs         []string
	Server       string
	Interface    string
	Mock         bool
	Refresh      time.Duration
	DoLog        bool
	AvoidUnicode bool
}

var options Options

type BlackHoleWriter struct {
}

func (w *BlackHoleWriter) Write(p []byte) (n int, err error) {
	err = errors.New("black hole writer")
	return
}

const (
	interfaceSimple   = "simple"
	interfaceAdvanced = "advanced"
)

var version *bool
var Version = "undefined"

func init() {
	jobs := flag.String("jobs", "", "CSV of all jobs on the server you want to track")
	doLog := flag.Bool("doLog", false, "Make a log of program execution")
	server := flag.String("server", "http://jenkins", "URL of the Jenkins server")
	intf := flag.String("interface", "advanced", "What interface should be used: console, advanced")
	mock := flag.Bool("mock", false, "Use mocked data to see how program behaves")
	refresh := flag.Duration("refresh", 15*time.Second, "How often to refresh Jenkins status")
	avoidUnicode := flag.Bool("avoidUnicode", false, "Will avoid usage of Unicode characters in terminal. V will mean Success, X will mean Failure, B will mean building")
	version = flag.Bool("version", false, "Get application version")
	flag.Parse()
	options = Options{
		Jobs:         strings.Split(*jobs, ","),
		Server:       *server,
		Interface:    *intf,
		Mock:         *mock,
		Refresh:      *refresh,
		DoLog:        *doLog,
		AvoidUnicode: *avoidUnicode,
	}
}

func mainLoop(feedbackChannel chan Command, ui *View) {
	if options.DoLog {
		fmt.Println("using " + filepath.Join(filepath.Dir(os.Args[0]), "jenkins_ping.log"))
		logFile, err := os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "jenkins_ping.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	} else {
		log.SetOutput(&BlackHoleWriter{})
	}
	var api Api
	if options.Mock {
		api = &MockApi{}
	} else {
		api = &JenkinsApi{
			ServerLocation: options.Server,
		}
	}
	controller := Controller{
		View:      *ui,
		API:       api,
		KnownJobs: options.Jobs,
	}
	ticker := time.NewTicker(options.Refresh)
	firstRun := make(chan bool, 1)
	firstRun <- true
	for {
		select {
		case x := <-feedbackChannel:
			log.Printf("Received: %q\n", x)
			switch x.group {
			case CmdShutdownGroup:
				log.Println("Bye!")
				ticker.Stop()
				return
			case CmdCloseGroup:
				controller.RemoveModals()
			case CmdShowHelpGroup:
				controller.ShowHelp()
			case CmdOpenCurrentJobGroup:
				controller.VisitCurrentJob(x.job)
			case CmdOpenPreviousJobGroup:
				controller.VisitPreviousJob(x.job)
			case CmdTestsForJobGroup:
				controller.ShowTests(x.job)
			}
		case <-ticker.C:
			controller.RefreshNodeInformation()
		case <-firstRun:
			controller.RefreshNodeInformation()
		}
	}
}

func main() {
	if *version {
		fmt.Printf("jenkins_ping version: %v\n", Version)
		return
	}
	var feedbackChannel = make(chan Command)
	var ui View
	var err error
	switch options.Interface {
	case interfaceSimple:
		ui, err = NewConsoleInterface(feedbackChannel)
	case interfaceAdvanced:
		ui, err = NewCUIInterface(feedbackChannel)
	}
	if err != nil {
		log.Println("Failure to activate advanced interface", err)
		ui, err = NewConsoleInterface(feedbackChannel)
	}
	if err != nil {
		log.Fatal("Failure to boot interface", err)
	} else {
		mainLoop(feedbackChannel, &ui)
	}
}
