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

// JenkinsAPIRoot carries information about a single Jenkins API root (server or a folder in a server)
// with all jobs under that endpoint
type JenkinsAPIRoot struct {
	API    jenkins.API
	Server string
	Jobs   []string
}

// Controller is a class that is a backend per-server notification source.
// It is able to communicate changes detected in state of the Jenkins server back to the View.
type Controller struct {
	View  view.View
	APIs  []JenkinsAPIRoot
	state model.State
}

// RefreshNodeInformation will start Jenkins API visiting and send updates to the view
// FIXME: it is not clear enough to know job names - you need server name as well
func (controller *Controller) RefreshNodeInformation(knownJobs []string) {
	log.Println("Controller: RefreshNodeInformation")
	state := &controller.state
	state.Error = nil
	for _, endpoint := range controller.APIs {
		resultFromJenkins, err := endpoint.API.GetKnownJobs()
		if err != nil {
			log.Printf("Error state for partial update: %v", err)
			break
		} else {
			jobStates, err := controller.explainProperStates(&endpoint, resultFromJenkins)
			if err != nil {
				log.Printf("Error state for partial update: %v", err)
				break
			}
			if len(jobStates) != 0 {
				for _, serverState := range jobStates {
					found := false
					for i, modelState := range state.JobStates {
						if modelState.JobName == serverState.JobName {
							state.JobStates[i] = *serverState
							found = true
							break
						}
					}
					if !found {
						state.JobStates = append(state.JobStates, *serverState)
					}
				}
			}
		}
	}
	controller.updateView()
}

// RefreshAllNodeInformation is going to visit all the servers and all the jobs
// and update the state
func (controller *Controller) RefreshAllNodeInformation() {
	log.Println("Controller: RefreshAllNodeInformation")
	state := &controller.state
	state.Error = nil
	state.JobStates = make([]model.JobState, 0)
	for _, endpoint := range controller.APIs {
		resultFromJenkins, err := endpoint.API.GetKnownJobs()
		if err != nil {
			log.Printf("Error state: %v", err)
			state.Error = err
			state.JobStates = make([]model.JobState, 0)
			break
		} else {
			jobStates, err := controller.explainProperStates(&endpoint, resultFromJenkins)
			if err != nil {
				state.Error = err
				state.JobStates = make([]model.JobState, 0)
				break
			}
			for _, jobState := range jobStates {
				state.JobStates = append(state.JobStates, *jobState)
			}
		}
	}
	controller.updateView()
}

func (controller *Controller) updateView() {
	if controller.View != nil {
		controller.View.PresentState(&controller.state)
	}
}

func (controller *Controller) explainProperStates(jenkinsAPIRoot *JenkinsAPIRoot, jenkinsAnswer *jenkins.Status) (jobStates []*model.JobState, err error) {
	if len(jenkinsAPIRoot.Jobs) == 1 && jenkinsAPIRoot.Jobs[0] == "" {
		for _, item := range jenkinsAnswer.JobBuildStatus {
			jobStates = append(jobStates, &model.JobState{
				JobName:       item.Name,
				Server:        jenkinsAPIRoot.Server,
				PreviousState: model.BuildStatusFromColor(item.Color),
			})
		}
	} else {
		for _, jobWeCareAbout := range jenkinsAPIRoot.Jobs {
			for _, item := range jenkinsAnswer.JobBuildStatus {
				if jobWeCareAbout == item.Name {
					jobStates = append(jobStates, &model.JobState{
						JobName:       item.Name,
						Server:        jenkinsAPIRoot.Server,
						PreviousState: model.BuildStatusFromColor(item.Color),
					})
				}
			}
		}
	}

	for ind := range jobStates {
		iterState := jobStates[ind]
		status, err := jenkinsAPIRoot.API.GetCurrentStatus(iterState.JobName)
		if err == nil {
			iterState.CausesFriendly = jenkinsAPIRoot.API.CausesFriendly(status)
			iterState.CulpritsFriendly = jenkinsAPIRoot.API.CausesOfPreviousFailuresFriendly(iterState.JobName)
			iterState.Building = status.Building
			iterState.Time = controller.explainTime(*status)
		} else {
			iterState.Error = err
		}
	}
	if len(jobStates) == 0 {
		err = fmt.Errorf("No jobs from %+v matched amongst following available jobs: %v", jenkinsAPIRoot, jenkinsAnswer)
	}
	return
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
	if api, ok := controller.apiForState(id); ok {
		controller.visitURL(api.GetLastBuildURLForJob(controller.state.JobStates[id].JobName))
	}
}

// VisitPreviousJob will open the browser and direct you to the url where last completed build for a certain job will be shown
func (controller *Controller) VisitPreviousJob(id int) {
	if api, ok := controller.apiForState(id); ok {
		controller.visitURL(api.GetLastCompletedBuildURLForJob(controller.state.JobStates[id].JobName))
	}
}

func (controller *Controller) apiForState(id int) (jenkins.API, bool) {
	if id >= len(controller.state.JobStates) {
		log.Printf("Unsupported index (out of bounds of known jobs): %v (max is %v)", id, len(controller.state.JobStates)-1)
	} else {
		state := controller.state.JobStates[id]
		for _, serverEndpoint := range controller.APIs {
			if serverEndpoint.Server == state.Server {
				return serverEndpoint.API, true
			}
		}
	}
	return nil, false
}

func (controller *Controller) visitURL(url string) {
	if err := open.Run(url); err != nil {
		log.Printf("Could not open URL %s!, err: %v", url, err)
	}
}

// ShowTests will ask Jenkins for failed tests in last execution of a certain job and update view with that info
func (controller *Controller) ShowTests(id int) {
	log.Println("Controller: ShowTests")
	if api, ok := controller.apiForState(id); ok {
		failedTests, err := api.GetFailedTestList(controller.state.JobStates[id].JobName)
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
