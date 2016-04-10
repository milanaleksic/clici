package jenkins

import (
	"fmt"
	"math/rand"
	"time"
)

// MockAPI is a mock implementation of the Jenkins API to be used in simple UI testing.
// This API doesn't visit any server and thus randomly generates data
type MockAPI struct {
}

var randomCauses = []string{
	"milan",
	"fred",
	"johnny",
	"unknown",
}

// GetKnownJobs is a MOCK for call that represents API which gives back list of all known jobs
func (api *MockAPI) GetKnownJobs() (resultFromJenkins *Status, err error) {
	resultFromJenkins = &Status{}
	resultFromJenkins.JobBuildStatus = make([]JobBuildStatus, 0)
	for i := 0; i < 12; i++ {
		var color string
		switch rand.Intn(3) {
		case 0:
			color = "blue"
		case 1:
			color = "red"
		case 2:
			color = "aborted"
		}
		resultFromJenkins.JobBuildStatus = append(resultFromJenkins.JobBuildStatus, JobBuildStatus{
			Name:  fmt.Sprintf("a_test_job_long_name%v", i),
			Color: color,
		})
	}
	return resultFromJenkins, nil
}

// GetCurrentStatus is a MOCK for call that returns current state for a particular job
func (api *MockAPI) GetCurrentStatus(job string) (status *JobStatus, err error) {
	var culprits = make([]Culprit, 0)
	for i := 0; i < rand.Intn(5); i++ {
		culprits = append(culprits, Culprit{
			FullName: randomCauses[rand.Intn(len(randomCauses))],
		})
	}
	var causes = make([]Cause, 0)
	for i := 0; i < rand.Intn(5); i++ {
		causes = append(causes, Cause{
			//TODO: mock also causes chain here
			UserID: randomCauses[rand.Intn(len(randomCauses))],
		})
	}
	result := &JobStatus{
		Building:          rand.Intn(2) == 0,
		EstimatedDuration: int64(rand.Intn(300000)),
		Timestamp:         time.Now().UnixNano()/1000/1000 - int64(rand.Intn(300000)),
		Culprits:          culprits,
		Actions: []Action{
			Action{
				Causes: causes,
			},
		},
	}
	return result, nil
}

// CausesFriendly is a MOCK for call that takes a known job status and finds people ("causes") that caused it to start,
// returning a CSV list of people.
func (api *MockAPI) CausesFriendly(status *JobStatus) string {
	set := make(map[string]bool, 0)
	for i := 0; i < rand.Intn(5); i++ {
		set[randomCauses[rand.Intn(len(randomCauses))]] = true
	}
	return joinKeysInCsv(set)
}

// CausesOfPreviousFailuresFriendly is a MOCK for call that finds reasons why a particular job previously fail,
// returning a CSV list of people who caused it
func (api *MockAPI) CausesOfPreviousFailuresFriendly(job string) string {
	return api.CausesFriendly(&JobStatus{})
}

// GetLastBuildURLForJob is a MOCK for call that will create URL towards a page with LAST job execution result for a particular job
func (api *MockAPI) GetLastBuildURLForJob(job string) string {
	return fmt.Sprintf("http://mock_jenkins/job/%v/lastBuild/", job)
}

// GetLastCompletedBuildURLForJob is a MOCK for call that will create URL towards a page with LAST COMPLETED job execution result for a particular job
func (api *MockAPI) GetLastCompletedBuildURLForJob(job string) string {
	return fmt.Sprintf("http://mock_jenkins/job/%v/lastCompletedBuild/", job)
}

// GetFailedTestList is a MOCK for call that will return list of test cases that failed in a LAST FAILED job execution
func (api *MockAPI) GetFailedTestList(job string) (testCaseResult []TestCase, err error) {
	var set []TestCase
	var randomTests = []string{
		"test1",
		"test2",
		"test3",
		"test4",
	}
	for i := 0; i < rand.Intn(5); i++ {
		aCase := TestCase{
			ClassName: randomTests[rand.Intn(len(randomTests))],
			Name:      randomTests[rand.Intn(len(randomTests))],
			Status:    "FAILED",
		}
		set = append(set, aCase)
	}
	return
}
