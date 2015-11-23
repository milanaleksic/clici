package main

import (
	"time"
	"fmt"
	"strings"
	"log"
	"github.com/skratchdot/open-golang/open"
)

type State struct {
	JobStates []JobState
	Error     error
}

func (state *State) MaxLengthOfName() (lengthForJobNames int) {
	lengthForJobNames = 10
	for _, jobState := range state.JobStates {
		if len(jobState.JobName) > lengthForJobNames {
			lengthForJobNames = len(jobState.JobName)
		}
	}
	return
}

type BuildStatus byte

const (
	Success BuildStatus = iota
	Failure
	Unknown
)

type JobState struct {
	JobName          string
	CulpritsFriendly string
	PreviousState    BuildStatus
	Error            error
	CausesFriendly   string
	Building         bool
	Time             string
}

type Controller struct {
	KnownJobs []string
	View      View
	API       Api
	state     *State
}

func (controller *Controller) RefreshNodeInformation() {
	state := &State{}
	resultFromJenkins, err := controller.API.GetRunningJobs()
	if err != nil {
		log.Printf("Error state: %v", err)
		state.Error = err
	} else {
		controller.explainProperState(resultFromJenkins, state)
	}
	controller.state = state
	controller.View.PresentState(state)
}

func (controller *Controller) explainProperState(resultFromJenkins *JenkinsStatus, state *State) {
	state.JobStates = make([]JobState, 0)
	if len(controller.KnownJobs) == 1 && controller.KnownJobs[0] == "" {
		for _, item := range resultFromJenkins.JobBuildStatus {
			state.JobStates = append(state.JobStates, JobState{
				JobName: item.Name,
				PreviousState: controller.previousStateFromColor(item.Color),
			})
		}
	} else {
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
	}

	for ind, _ := range state.JobStates {
		iterState := &state.JobStates[ind]
		status, err := controller.API.GetCurrentStatus(iterState.JobName)
		if err == nil {
			iterState.CausesFriendly = controller.API.CausesFriendly(status)
			iterState.CulpritsFriendly = copyCulprits(status)
			iterState.Building = status.Building
			iterState.Time = controller.ExplainTime(*status)
		} else {
			iterState.Error = err
		}
	}
}

func copyCulprits(status *JobStatus) (culpritsCsv string) {
	set := make(map[string]bool, 0)
	for _, culprit := range status.Culprits {
		set[culprit.FullName] = true
	}
	return joinKeysInCsv(set)
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

func (controller *Controller) VisitPageBehindId(id string) {
	if int(idtoindex(id[0])) >= len(controller.state.JobStates) {
		log.Printf("Unsupported index (out of bounds of known jobs): %v->%v (max is %v)", id, idtoindex(id[0]), len(controller.state.JobStates) - 1)
	} else {
		url := controller.API.GetLastBuildUrlForJob(controller.state.JobStates[idtoindex(id[0])].JobName)
		open.Run(url)
	}
}
