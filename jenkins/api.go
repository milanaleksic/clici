package jenkins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// API is defining known and supported calls towards a Jenkins server
type API interface {
	GetKnownJobs() (resultFromJenkins *Status, err error)
	GetCurrentStatus(job string) (status *JobStatus, err error)
	CausesFriendly(status *JobStatus) string
	CausesOfPreviousFailuresFriendly(job string) string
	GetLastBuildURLForJob(job string) string
	GetLastCompletedBuildURLForJob(job string) string
	GetFailedTestList(job string) (testCaseResult []TestCase, err error)
}

// ServerAPI is a real-life implementation of the API which connects to a real Jenkins server.
// Use the given "ServerLocation" field to set the location of the server.
type ServerAPI struct {
	ServerLocation string
	cachedStatuses map[string](*JobStatus)
}

// GetLastBuildURLForJob will create URL towards a page with LAST job execution result for a particular job
func (api *ServerAPI) GetLastBuildURLForJob(job string) string {
	return fmt.Sprintf("%v/job/%v/lastBuild/", api.ServerLocation, job)
}

// GetLastCompletedBuildURLForJob will create URL towards a page with LAST COMPLETED job execution result for a particular job
func (api *ServerAPI) GetLastCompletedBuildURLForJob(job string) string {
	return fmt.Sprintf("%v/job/%v/lastCompletedBuild/", api.ServerLocation, job)
}

// DTOs
// 1. DON'T TRY to use camelCase in DTOs, json unmarshaller doesn't see it!
// 2. DON'T TRY to put space between ":" and "\"", unmarshaller doesn't see it (sometimes)!

// JobStatus contains a parsed Jenkins server response about a single job result status
type JobStatus struct {
	ID                string    `json:"id"`
	Result            string    `json:"result"`
	Building          bool      `json:"building"`
	Actions           []action  `json:"actions"`
	Culprits          []culprit `json:"culprits"`
	EstimatedDuration int64     `json:"estimatedDuration"`
	Timestamp         int64     `json:"timestamp"`
}

type culprit struct {
	FullName string `json:"fullName"`
}

type action struct {
	Causes []cause `json:"causes"`
}

type cause struct {
	UserID           string `json:"userId"`
	UpstreamBuild    int    `json:"upstreamBuild"`
	UpstreamProject  string `json:"upstreamProject"`
	ShortDescription string `json:"shortDescription"`
}

// GetCurrentStatus returns current state for a particular job
func (api *ServerAPI) GetCurrentStatus(job string) (status *JobStatus, err error) {
	return api.getStatusForJob(job, "lastBuild")
}

func (api *ServerAPI) getStatusForJob(job string, id string) (status *JobStatus, err error) {
	possibleCacheKey := fmt.Sprintf("%s-%s", job, id)
	if id != "lastBuild" && id != "lastCompletedBuild" {
		if api.cachedStatuses == nil {
			api.cachedStatuses = make(map[string](*JobStatus), 0)
		}
		if cachedValue, ok := api.cachedStatuses[possibleCacheKey]; ok {
			log.Println("Using from cache: ", possibleCacheKey)
			return cachedValue, nil
		}
	}
	link := fmt.Sprintf("%v/job/%v/%v/api/json?tree=id,result,timestamp,estimatedDuration,building,culprits[fullName],actions[causes[userId,upstreamBuild,upstreamProject,shortDescription]]",
		api.ServerLocation, job, id)
	log.Printf("Visiting %v", link)
	resp, err := http.Get(link)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	result := &JobStatus{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err == nil && id != "lastBuild" && id != "lastCompletedBuild" {
		api.cachedStatuses[possibleCacheKey] = result
	}
	return result, nil
}

// Status represents API response for list of currently known jobs in the Jenkins Server.
type Status struct {
	JobBuildStatus []jobBuildStatus `json:"jobs"`
}

type jobBuildStatus struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// GetKnownJobs represents API which gives back list of all known jobs in the Jenkins Server, and their last known
// (or current, if job is running) state
func (api *ServerAPI) GetKnownJobs() (resultFromJenkins *Status, err error) {
	resp, err := http.Get(fmt.Sprintf("%v/api/json?tree=jobs[name,color]", api.ServerLocation))
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	resultFromJenkins = &Status{}
	err = json.NewDecoder(resp.Body).Decode(&resultFromJenkins)
	return resultFromJenkins, nil
}

// CausesFriendly takes a known job status and finds people ("causes") that caused it to start,
// returning a CSV list of people.
// It might need to visit server again in case it has to follow casual chain
func (api *ServerAPI) CausesFriendly(status *JobStatus) string {
	set := make(map[string]bool, 0)
	for _, culprit := range status.Culprits {
		set[culprit.FullName] = true
	}
	api.addActionIdsToSet(set, status.Actions)
	return joinKeysInCsv(set)
}

// CausesOfPreviousFailuresFriendly finds reasons why a particular job previously fail,
// returning a CSV list of people who caused it
func (api *ServerAPI) CausesOfPreviousFailuresFriendly(name string) string {
	set := make(map[string]bool, 0)
	id := "lastCompletedBuild"
	for {
		statusIterator, err := api.getStatusForJob(name, id)
		if err != nil {
			log.Println("Could not fetch causes: ", err)
			return "?"
		}
		if statusIterator.Result == "SUCCESS" {
			break
		}
		api.addActionIdsToSet(set, statusIterator.Actions)
		api.addCulpritIdsToSet(set, statusIterator.Culprits)
		currentID, err := strconv.Atoi(statusIterator.ID)
		if err != nil {
			log.Println("Could not parse number: ", statusIterator.ID, err)
			return "?"
		}
		id = strconv.Itoa(currentID - 1)
	}
	return joinKeysInCsv(set)
}

func (api *ServerAPI) addCulpritIdsToSet(set map[string]bool, culprits []culprit) {
	for _, culprit := range culprits {
		set[culprit.FullName] = true
	}
	return
}

func (api *ServerAPI) addActionIdsToSet(set map[string]bool, actions []action) {
	for _, action := range actions {
		for _, cause := range action.Causes {
			if cause.UserID != "" {
				set[cause.UserID] = true
			} else if cause.UpstreamBuild != 0 && cause.UpstreamProject != "" {
				new, err := api.addCauses(cause.UpstreamProject, cause.UpstreamBuild)
				if err != nil {
					log.Printf("Could not catch causes: %v\n", err)
				} else {
					for _, new := range new {
						set[new] = true
					}
				}
			}
		}
	}
	return
}

func (api *ServerAPI) addCauses(upstreamProject string, upstreamBuild int) (target []string, err error) {
	status, err := api.getStatusForJob(upstreamProject, strconv.Itoa(upstreamBuild))
	target = make([]string, 0)
	for _, action := range status.Actions {
		for _, cause := range action.Causes {
			if cause.UserID != "" {
				target = append(target, cause.UserID)
			} else if cause.UpstreamBuild != 0 && cause.UpstreamProject != "" {
				new, err2 := api.addCauses(cause.UpstreamProject, cause.UpstreamBuild)
				if err2 != nil {
					err = err2
					return
				}
				target = append(target, new...)
			} else if cause.ShortDescription == "Started by an SCM change" {
				for _, culprit := range status.Culprits {
					target = append(target, culprit.FullName)
				}
			}
		}
	}
	return
}

type testCaseResult struct {
	Suites []testSuite `json:"suites"`
}

type testSuite struct {
	Cases []TestCase `json:"cases"`
}

// TestCase depicts part of Jenkins API and identifies which particular test case failed while running a job
type TestCase struct {
	ClassName string `json:"className"`
	Name      string `json:"name"`
	Status    string `json:"status"`
}

// GetFailedTestList will return list of test cases that failed in a LAST FAILED job execution
func (api *ServerAPI) GetFailedTestList(job string) (results []TestCase, err error) {
	link := fmt.Sprintf("%v/job/%s/lastFailedBuild/testReport/api/json?tree=suites[cases[className,name,status]]", api.ServerLocation, job)
	log.Printf("Visiting %s\n", link)
	resp, err := http.Get(link)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	var received testCaseResult
	err = json.NewDecoder(resp.Body).Decode(&received)
	if err != nil {
		return
	}

	results = make([]TestCase, 0)
	for _, suite := range received.Suites {
		for _, aCase := range suite.Cases {
			if aCase.Status != "PASSED" && aCase.Status != "SKIPPED" {
				results = append(results, aCase)
			}
		}
	}
	return
}
