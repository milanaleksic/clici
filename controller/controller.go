package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/model"
	"github.com/milanaleksic/clici/view"
	"github.com/skratchdot/open-golang/open"
)

// Controller is a class that is a backend per-server notification source.
// It is able to communicate changes detected in state of the Jenkins server back to the View.
type Controller struct {
	View      view.View
	API       jenkins.API
	state     model.State
}

// RefreshNodeInformation will start Jenkins API visiting and send updates to the view
func (controller *Controller) RefreshNodeInformation(knownJobs []string) {
	log.Println("Controller: RefreshNodeInformation")
	state := &controller.state
	resultFromJenkins, err := controller.API.GetKnownJobs()
	if err != nil {
		log.Printf("Error state: %v", err)
		state.Error = err
	} else {
		controller.explainProperState(knownJobs, resultFromJenkins)
	}
	controller.updateView()
}

func (controller *Controller) updateView() {
	if controller.View != nil {
		controller.View.PresentState(&controller.state)
	}
}

func (controller *Controller) explainProperState(knownJobs []string, resultFromJenkins *jenkins.Status) {
	state := &controller.state
	state.Error = nil
	state.JobStates = make([]model.JobState, 0)
	if len(knownJobs) == 1 && knownJobs[0] == "" {
		for _, item := range resultFromJenkins.JobBuildStatus {
			state.JobStates = append(state.JobStates, model.JobState{
				JobName:       item.Name,
				PreviousState: model.BuildStatusFromColor(item.Color),
			})
		}
	} else {
		for _, jobWeCareAbout := range knownJobs {
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
	if len(state.JobStates) == 0 {
		state.Error = fmt.Errorf("No jobs from %v matched amongst following available jobs: %v", knownJobs, resultFromJenkins)
	}
}

func (controller *Controller) explainTime(status jenkins.JobStatus) string {
	secLeft := status.EstimatedDuration/1000 - (time.Now().UnixNano()/1000/1000-status.Timestamp)/1000
	if status.Building {
		if secLeft >= 0 {
			return fmt.Sprintf("%v min more", secLeft/60)
		}
		return fmt.Sprintf("%v min longer than expected", -secLeft/60)
	}
	return humanize.Time(time.Now().Add(time.Duration(secLeft) * time.Second))
}

// VisitCurrentJob will open the browser and direct you to the url where last build for a certain job will be shown
func (controller *Controller) VisitCurrentJob(id int) {
	controller.visitURL(id, controller.API.GetLastBuildURLForJob)
}

// VisitPreviousJob will open the browser and direct you to the url where last completed build for a certain job will be shown
func (controller *Controller) VisitPreviousJob(id int) {
	controller.visitURL(id, controller.API.GetLastCompletedBuildURLForJob)
}

func (controller *Controller) visitURL(id int, urlFromJobName func(job string) string) {
	if id >= len(controller.state.JobStates) {
		log.Printf("Unsupported index (out of bounds of known jobs): %v (max is %v)", id, len(controller.state.JobStates)-1)
	} else {
		url := urlFromJobName(controller.state.JobStates[id].JobName)
		if err := open.Run(url); err != nil {
			log.Printf("Could not open URL %s!, err: %v", url, err)
		}
	}
}

// ShowTests will ask Jenkins for failed tests in last execution of a certain job and update view with that info
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

// ShowHelp will update view with a flag to show help
func (controller *Controller) ShowHelp() {
	log.Println("Controller: ShowHelp")
	controller.state.ShowHelp = true
	controller.updateView()
}

// RemoveModals will update view with a flag to remove all modal dialogs
func (controller *Controller) RemoveModals() {
	log.Println("Controller: RemoveModals")
	controller.state.ShowHelp = false
	controller.state.Error = nil
	controller.state.FailedTests = nil
	controller.updateView()
}
