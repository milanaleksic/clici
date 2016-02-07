package model

import (
	"log"
	"strings"
)

// State is the program state model that mutates based on the Jenkins server state
// and based on human interaction with the view
type State struct {
	JobStates   []JobState
	FailedTests []string
	Error       error
	ShowHelp    bool
}

// BuildStatus is a model way of representing a status of a certain job in Jenkins
type BuildStatus byte

// BuildStatusFromColor returns a model representation of the build status based on known Jenkins
// code statuses
func BuildStatusFromColor(color string) BuildStatus {
	if strings.Index(color, "blue") == 0 {
		return Success
	} else if strings.Index(color, "red") == 0 {
		return Failure
	} else {
		log.Printf("Unknown color: %v\n", color)
		return Unknown
	}
}

const (
	// Success means that job has finished with full success
	Success BuildStatus = iota
	// Failure means that job has failed (either tests have failed of there was some kind of wrong processing)
	Failure
	// Unknown job state means that this application is not able to deduce job state
	Unknown
)

// JobState is full representation of job state in Jenkins, with all known data program can extract at this time
type JobState struct {
	JobName          string
	CulpritsFriendly string
	CausesFriendly   string
	Time             string
	Error            error
	PreviousState    BuildStatus
	Building         bool
}
