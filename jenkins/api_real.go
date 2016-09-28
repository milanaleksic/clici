package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	lastCompletedBuild = "lastCompletedBuild"
	lastBuild          = "lastBuild"
	sizeOfSuffix       = 2048
)

var (
	errStatusPageNotFound = errors.New("Not Found")
)

// ServerAPI is a real-life implementation of the API which connects to a real Jenkins server.
// Use the given "ServerLocation" field to set the location of the server.
type ServerAPI struct {
	ServerLocation string
	cachedStatuses map[string](*JobStatus)
}

// GetLastBuildURLForJob will create URL towards a page with LAST job execution result for a particular job
func (api *ServerAPI) GetLastBuildURLForJob(job string) string {
	return fmt.Sprintf("%v/job/%v/%v/", api.ServerLocation, job, lastBuild)
}

// GetLastCompletedBuildURLForJob will create URL towards a page with LAST COMPLETED job execution result for a particular job
func (api *ServerAPI) GetLastCompletedBuildURLForJob(job string) string {
	return fmt.Sprintf("%v/job/%v/%v/", api.ServerLocation, job, lastCompletedBuild)
}

// GetCurrentStatus returns current state for a particular job
func (api *ServerAPI) GetCurrentStatus(job string) (*JobStatus, error) {
	return api.GetStatusForJob(job, lastBuild)
}

// GetStatusForJob returns a status of a specific job run
func (api *ServerAPI) GetStatusForJob(job string, id string) (*JobStatus, error) {
	possibleCacheKey := fmt.Sprintf("%s-%s", job, id)
	if id != lastBuild && id != lastCompletedBuild {
		if api.cachedStatuses == nil {
			api.cachedStatuses = make(map[string](*JobStatus), 0)
		}
		if cachedValue, ok := api.cachedStatuses[possibleCacheKey]; ok {
			log.Println("Using from cache: ", possibleCacheKey)
			return cachedValue, nil
		}
	}
	link := fmt.Sprintf("%v/job/%v/%v/api/json?tree=id,result,timestamp,estimatedDuration,building,culprits[fullName],actions[causes[userId,upstreamBuild,upstreamProject,shortDescription]],changeSets[items[author[fullName]]]",
		api.ServerLocation, job, id)
	log.Printf("Visiting %v", link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, errStatusPageNotFound
	}
	defer func() { _ = resp.Body.Close() }()
	result := &JobStatus{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err == nil && id != lastBuild && id != lastCompletedBuild {
		api.cachedStatuses[possibleCacheKey] = result
	}
	return result, nil
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
	api.addActionIdsToSet(set, status)
	return joinKeysInCsv(set)
}

// CausesOfFailuresFriendly finds reasons why a particular job which previously failed,
// returning a CSV list of people who caused it
func (api *ServerAPI) CausesOfFailuresFriendly(name, id string) string {
	set := make(map[string]bool, 0)
	var visitsToServerAllowed = 20
	for {
		currentID, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("Could not parse number: %s, reason: %+v; will give up from fetching further causes\n", id, err)
			break
		}
		visitsToServerAllowed--
		if visitsToServerAllowed <= 0 {
			log.Println("Maximum number of visits to Jenkins server reached, giving up from further changes")
			break
		}
		statusIterator, err := api.GetStatusForJob(name, id)
		if err != nil {
			if err == errStatusPageNotFound {
				id = strconv.Itoa(currentID - 1)
				continue
			}
			log.Println("Could not fetch causes: ", err)
			break
		}
		if statusIterator.Result == "SUCCESS" || statusIterator.Result == "FIXED" {
			break
		}
		log.Printf("Got actions %+v and culprits %+v from job=%s, id=%s\n", statusIterator.Actions, statusIterator.Culprits, name, id)
		api.addActionIdsToSet(set, statusIterator)
		api.addCulpritIdsToSet(set, statusIterator.Culprits)
		api.addChangeSetsToSet(set, statusIterator.ChangeSets)
		id = strconv.Itoa(currentID - 1)
	}
	return joinKeysInCsv(set)
}

// CausesOfPreviousFailuresFriendly finds reasons why the last execution of this job failed,
// returning a CSV list of people who caused it
func (api *ServerAPI) CausesOfPreviousFailuresFriendly(name string) string {
	return api.CausesOfFailuresFriendly(name, lastCompletedBuild)
}

func (api *ServerAPI) addCulpritIdsToSet(set map[string]bool, culprits []Culprit) {
	for _, culprit := range culprits {
		set[culprit.FullName] = true
	}
	return
}

func (api *ServerAPI) addChangeSetsToSet(set map[string]bool, changeSets []ChangeSet) {
	for _, changeSet := range changeSets {
		for _, changeSetItem := range changeSet.Items {
			set[changeSetItem.Author.FullName] = true
		}
	}
}

func (api *ServerAPI) addActionIdsToSet(set map[string]bool, status *JobStatus) {
	for _, action := range status.Actions {
		for _, cause := range action.Causes {
			if cause.UserID != "" {
				set[cause.UserID] = true
			} else if cause.UpstreamBuild != 0 && cause.UpstreamProject != "" {
				if err := api.addCauses(set, cause.UpstreamProject, cause.UpstreamBuild); err != nil {
					log.Printf("Could not catch causes: %v\n", err)
				}
			} else if cause.ShortDescription == "Started by an SCM change" {
				api.addCulpritIdsToSet(set, status.Culprits)
			} else if strings.HasPrefix(cause.ShortDescription, "commit notification") {
				api.addChangeSetsToSet(set, status.ChangeSets)
			}
		}
	}
	return
}

func (api *ServerAPI) addCauses(set map[string]bool, upstreamProject string, upstreamBuild int) error {
	status, err := api.GetStatusForJob(upstreamProject, strconv.Itoa(upstreamBuild))
	if err != nil {
		return err
	}
	api.addActionIdsToSet(set, status)
	return nil
}

// GetFailedTestListFor will return list of test cases that failed in a particular job execution
func (api *ServerAPI) GetFailedTestListFor(job, id string) (results []TestCase, err error) {
	link := fmt.Sprintf("%v/job/%s/%s/testReport/api/json?tree=suites[cases[className,name,status,errorStackTrace]]", api.ServerLocation, job, id)
	log.Printf("Visiting %s\n", link)
	resp, err := http.Get(link)
	if err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 404 {
		return nil, errors.New("no test report found")
	}
	var received TestCaseResult
	err = json.NewDecoder(resp.Body).Decode(&received)
	if err != nil {
		return
	}

	results = make([]TestCase, 0)
	for _, suite := range received.Suites {
		for _, aCase := range suite.Cases {
			if aCase.Status != "PASSED" && aCase.Status != "SKIPPED" && aCase.Status != "FIXED" {
				results = append(results, aCase)
			}
		}
	}
	return
}

// GetFailedTestList will return list of test cases that failed in a LAST FAILED job execution
func (api *ServerAPI) GetFailedTestList(job string) ([]TestCase, error) {
	return api.GetFailedTestListFor(job, "lastFailedBuild")
}

func fetchSizeForLastLogLines(linkForSize string) (int, error) {
	resp, err := http.Head(linkForSize)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("could not fetch log size, statusCode=%d", resp.StatusCode)
	}
	textSize := resp.Header.Get("X-Text-Size")
	if textSize == "" {
		return 0, errors.New("size not received from server HEAD call")
	}

	return strconv.Atoi(textSize)
}

func fetchLinesForLastLogLines(link string, lineCount int) (response []string, err error) {
	respData, err := http.Get(link)
	if err != nil {
		return
	}
	defer func() { _ = respData.Body.Close() }()
	if respData.StatusCode != 200 {
		return nil, fmt.Errorf("not able to fetch console output: %d", respData.StatusCode)
	}
	data, err := ioutil.ReadAll(respData.Body)
	if err != nil {
		return nil, err
	}
	var dataAsString []string
	nl, endIter := 0, len(data)-1
	for i := endIter; i >= 0 && nl < lineCount; i-- {
		if data[i] == '\n' && i != endIter {
			nl++
			dataAsString = append(dataAsString, string(data[i+1:endIter]))
			endIter = i
		}
	}
	for i := 0; i < len(dataAsString)/2; i++ {
		dataAsString[i], dataAsString[len(dataAsString)-i-1] = dataAsString[len(dataAsString)-i-1], dataAsString[i]
	}
	return dataAsString, nil
}

// GetLastLogLines returns lineCount lines from the console output of a job run
func (api *ServerAPI) GetLastLogLines(job, id string, lineCount int) (response []string, err error) {
	linkForSize := fmt.Sprintf("%v/job/%s/%s/logText/progressiveHtml", api.ServerLocation, job, id)
	size, err := fetchSizeForLastLogLines(linkForSize)
	if err != nil {
		return nil, err
	}
	return fetchLinesForLastLogLines(fmt.Sprintf("%s?start=%d", linkForSize, size-sizeOfSuffix), lineCount)
}
