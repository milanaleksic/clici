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

	"github.com/milanaleksic/jenkins_ping/jenkins"
	"github.com/milanaleksic/jenkins_ping/view"
)

type appOptions struct {
	Jobs      []string
	Server    string
	Interface string
	Refresh   time.Duration
	Mock      bool
	DoLog     bool
}

var options appOptions

type blackHoleWriter struct {
}

func (w *blackHoleWriter) Write(p []byte) (n int, err error) {
	err = errors.New("black hole writer")
	return
}

const (
	interfaceSimple   = "simple"
	interfaceAdvanced = "advanced"
)

var logFile *os.File

// Version holds the main version string which should be updated externally when building release
var Version = "undefined"
var version *bool

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
	options = appOptions{
		Jobs:      strings.Split(*jobs, ","),
		Server:    *server,
		Interface: *intf,
		Mock:      *mock,
		Refresh:   *refresh,
		DoLog:     *doLog,
	}
	view.AvoidUnicode = *avoidUnicode
}

func setupLog() {
	if options.DoLog {
		fmt.Println("using " + filepath.Join(filepath.Dir(os.Args[0]), "jenkins_ping.log"))
		var err error
		logFile, err = os.OpenFile(filepath.Join(filepath.Dir(os.Args[0]), "jenkins_ping.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		log.SetOutput(logFile)
	} else {
		log.SetOutput(&blackHoleWriter{})
	}
}

func getAPI() (api jenkins.API) {
	if options.Mock {
		api = &jenkins.MockAPI{}
	} else {
		api = &jenkins.ServerAPI{
			ServerLocation: options.Server,
		}
	}
	return
}

func main() {
	if *version {
		fmt.Printf("jenkins_ping version: %v\n", Version)
		return
	}
	setupLog()
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()
	var feedbackChannel = make(chan view.Command)
	var ui view.View
	var err error
	switch options.Interface {
	case interfaceSimple:
		ui, err = view.NewConsoleInterface(feedbackChannel)
	case interfaceAdvanced:
		ui, err = view.NewCUIInterface(feedbackChannel)
	}
	if err != nil || ui == nil {
		log.Println("Failure to activate advanced interface", err)
		ui, err = view.NewConsoleInterface(feedbackChannel)
	}
	if err != nil {
		log.Fatal("Failure to boot interface", err)
	} else {
		dispatcher := &dispatcher{
			feedbackChannel: feedbackChannel,
			controller: &controller{
				View:      ui,
				API:       getAPI(),
				KnownJobs: options.Jobs,
			},
		}
		dispatcher.mainLoop()
	}
}
