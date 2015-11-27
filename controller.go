package main

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"log"
	"strings"
	"time"
)

type State struct {
	JobStates   []JobState
	FailedTests []string
	Error       error
	ShowHelp    bool
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
	state     State
}

func (controller *Controller) RefreshNodeInformation() {
	log.Println("Controller: RefreshNodeInformation")
	state := controller.state
	resultFromJenkins, err := controller.API.GetRunningJobs()
	if err != nil {
		log.Printf("Error state: %v", err)
		state.Error = err
	} else {
		controller.explainProperState(resultFromJenkins)
	}
	controller.updateView()
}

func (controller *Controller) updateView() {
	controller.View.PresentState(&controller.state)
}

func (controller *Controller) explainProperState(resultFromJenkins *JenkinsStatus) {
	state := &controller.state
	state.JobStates = make([]JobState, 0)
	if len(controller.KnownJobs) == 1 && controller.KnownJobs[0] == "" {
		for _, item := range resultFromJenkins.JobBuildStatus {
			state.JobStates = append(state.JobStates, JobState{
				JobName:       item.Name,
				PreviousState: controller.previousStateFromColor(item.Color),
			})
		}
	} else {
		for _, jobWeCareAbout := range controller.KnownJobs {
			for _, item := range resultFromJenkins.JobBuildStatus {
				if jobWeCareAbout == item.Name {
					state.JobStates = append(state.JobStates, JobState{
						JobName:       item.Name,
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
			iterState.CulpritsFriendly = controller.API.CausesOfPreviousFailureFriendly(iterState.JobName)
			iterState.Building = status.Building
			iterState.Time = controller.explainTime(*status)
		} else {
			iterState.Error = err
		}
	}
}

func (controller *Controller) explainTime(status JobStatus) string {
	timeLeft := status.EstimatedDuration/1000/60 - (time.Now().UnixNano()/1000/1000-status.Timestamp)/1000/60
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

func (controller *Controller) VisitCurrentJob(id int) {
	if int(id) >= len(controller.state.JobStates) {
		log.Printf("Unsupported index (out of bounds of known jobs): %v (max is %v)", id, len(controller.state.JobStates)-1)
	} else {
		url := controller.API.GetLastBuildUrlForJob(controller.state.JobStates[id].JobName)
		open.Run(url)
	}
}

func (controller *Controller) VisitPreviousJob(id int) {
	if int(id) >= len(controller.state.JobStates) {
		log.Printf("Unsupported index (out of bounds of known jobs): %v (max is %v)", id, len(controller.state.JobStates)-1)
	} else {
		url := controller.API.GetLastCompletedBuildUrlForJob(controller.state.JobStates[id].JobName)
		open.Run(url)
	}
}

func (controller *Controller) ShowTests(id int) {
	log.Println("Controller: ShowTests")
	failedTests, err := controller.API.GetFailedTestList(controller.state.JobStates[id].JobName)
	if err != nil {
		log.Printf("Error state: %v", err)
		controller.state.Error = err
	} else {
		testNames := make([]string, len(failedTests))
		for i, failedTest := range failedTests {
			testNames[i] = fmt.Sprintf("%s %s", failedTest.ClassName, failedTest.Name)
		}
		controller.state.FailedTests = testNames
	}
	controller.updateView()
}

func (controller *Controller) ShowHelp() {
	log.Println("Controller: ShowHelp")
	controller.state.ShowHelp = true
	controller.updateView()
}

func (controller *Controller) RemoveModals() {
	log.Println("Controller: RemoveModals")
	controller.state.ShowHelp = false
	controller.state.Error = nil
	controller.state.FailedTests = nil
	controller.updateView()
}
