package main

import (
	"time"
	"fmt"
	"strings"
	"log"
)

type State struct {
	JobStates []JobState
	Error     error
}

type BuildStatus byte

const (
	Success BuildStatus = iota
	Failure
	Unknown
)

type JobState struct {
	JobName        string
	Culprits       []string
	PreviousState  BuildStatus
	Error          error
	CausesFriendly string
	Building       bool
	Time           string
}

type Controller struct {
	View      View
	API       JenkinsApi
	KnownJobs []string
}

func (controller *Controller) RefreshNodeInformation() {
	state := &State{}
	resultFromJenkins, err := controller.API.GetRunningJobs()
	if err != nil {
		state.Error = err
		return
	}
	state.JobStates = make([]JobState, 0)
	for _, jobWeCareAbout := range controller.KnownJobs {
		for _, item := range resultFromJenkins.JobBuildStatus {
			if jobWeCareAbout == item.Name {
				state.JobStates = append(state.JobStates, JobState{
					JobName: item.Name,
					PreviousState: controller.previousStateFromColor(item.Color),
				})
			}
		}
	}

	for ind, _ := range state.JobStates {
		iterState := &state.JobStates[ind]
		status, err := controller.API.GetCurrentStatus(iterState.JobName)
		if err == nil {
			iterState.CausesFriendly = controller.API.CausesFriendly(&status)
			iterState.Building = status.Building
			iterState.Time = controller.ExplainTime(status)
		} else {
			iterState.Error = err
		}
	}
	controller.View.PresentState(state)
}

func (controller *Controller) ExplainTime(status JobStatus) string {
	timeLeft := status.EstimatedDuration / 1000 / 60 - (time.Now().UnixNano() / 1000 / 1000 - status.Timestamp) / 1000 / 60
	if timeLeft >= 0 {
		return fmt.Sprintf("%v min more", timeLeft)
	} else {
		return fmt.Sprintf("%v min longer than expected", -timeLeft)
	}
}

func (controller *Controller) previousStateFromColor(color string) BuildStatus {
	if strings.Index(color, "blue") == 0 {
		return Success
	} else if strings.Index(color, "red") == 0 {
		return Failure
	} else {
		log.Printf("Unknown color: %v\n", color)
		return Unknown
	}
}