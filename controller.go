package main

import (
	"fmt"
	"log"
	"time"

	"github.com/milanaleksic/jenkins_ping/jenkins"
	"github.com/milanaleksic/jenkins_ping/model"
	"github.com/milanaleksic/jenkins_ping/view"
	"github.com/skratchdot/open-golang/open"
)

type controller struct {
	KnownJobs []string
	View      view.View
	API       jenkins.API
	state     model.State
}

func (controller *controller) RefreshNodeInformation() {
	log.Println("Controller: RefreshNodeInformation")
	state := &controller.state
	resultFromJenkins, err := controller.API.GetKnownJobs()
	if err != nil {
		log.Printf("Error state: %v", err)
		state.Error = err
	} else {
		controller.explainProperState(resultFromJenkins)
	}
	controller.updateView()
}

func (controller *controller) updateView() {
	controller.View.PresentState(&controller.state)
}

func (controller *controller) explainProperState(resultFromJenkins *jenkins.Status) {
	state := &controller.state
	state.Error = nil
	state.JobStates = make([]model.JobState, 0)
	if len(controller.KnownJobs) == 1 && controller.KnownJobs[0] == "" {
		for _, item := range resultFromJenkins.JobBuildStatus {
			state.JobStates = append(state.JobStates, model.JobState{
				JobName:       item.Name,
				PreviousState: model.BuildStatusFromColor(item.Color),
			})
		}
	} else {
		for _, jobWeCareAbout := range controller.KnownJobs {
			for _, item := range resultFromJenkins.JobBuildStatus {
				if jobWeCareAbout == item.Name {
					state.JobStates = append(state.JobStates, model.JobState{
						JobName:       item.Name,
						PreviousState: model.BuildStatusFromColor(item.Color),
					})
				}
			}
		}
	}

	for ind := range state.JobStates {
		iterState := &state.JobStates[ind]
		status, err := controller.API.GetCurrentStatus(iterState.JobName)
		if err == nil {
			iterState.CausesFriendly = controller.API.CausesFriendly(status)
			iterState.CulpritsFriendly = controller.API.CausesOfPreviousFailuresFriendly(iterState.JobName)
			iterState.Building = status.Building
			iterState.Time = controller.explainTime(*status)
		} else {
			iterState.Error = err
		}
	}
}

func (controller *controller) explainTime(status jenkins.JobStatus) string {
	timeLeft := status.EstimatedDuration/1000/60 - (time.Now().UnixNano()/1000/1000-status.Timestamp)/1000/60
	if status.Building {
		if timeLeft >= 0 {
			return fmt.Sprintf("%v min more", timeLeft)
		}
		return fmt.Sprintf("%v min longer than expected", -timeLeft)
	} else {
		if timeLeft < 0 {
			return fmt.Sprintf("%v min ago", -timeLeft)
		} else {
			return fmt.Sprintf("will have finished (hhgttg or bad data) %v min ago", timeLeft)
		}
	}
}

func (controller *controller) VisitCurrentJob(id int) {
	controller.visitURL(id, controller.API.GetLastBuildURLForJob)
}

func (controller *controller) VisitPreviousJob(id int) {
	controller.visitURL(id, controller.API.GetLastCompletedBuildURLForJob)
}

func (controller *controller) visitURL(id int, urlFromJobName func(job string) string) {
	if int(id) >= len(controller.state.JobStates) {
		log.Printf("Unsupported index (out of bounds of known jobs): %v (max is %v)", id, len(controller.state.JobStates)-1)
	} else {
		url := urlFromJobName(controller.state.JobStates[id].JobName)
		if err := open.Run(url); err != nil {
			log.Printf("Could not open URL %s!, err: %v", url, err)
		}
	}
}

func (controller *controller) ShowTests(id int) {
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

func (controller *controller) ShowHelp() {
	log.Println("Controller: ShowHelp")
	controller.state.ShowHelp = true
	controller.updateView()
}

func (controller *controller) RemoveModals() {
	log.Println("Controller: RemoveModals")
	controller.state.ShowHelp = false
	controller.state.Error = nil
	controller.state.FailedTests = nil
	controller.updateView()
}
