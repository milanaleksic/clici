package main

import (
	"strings"
	"math/rand"
	"time"
	"fmt"
)

type MockApi struct {
}

var random_causes = []string{
	"milan",
	"fred",
	"johnny",
	"unknown",
}

func (api *MockApi) GetRunningJobs() (resultFromJenkins *JenkinsStatus, err error) {
	resultFromJenkins = &JenkinsStatus{}
	resultFromJenkins.JobBuildStatus = make([]JobBuildStatus, 0)
	for i := 0; i < 12; i++ {
		var color string
		if rand.Intn(2) == 0 {
			color = "blue"
		} else {
			color = "red"
		}
		resultFromJenkins.JobBuildStatus = append(resultFromJenkins.JobBuildStatus, JobBuildStatus{
			Name: fmt.Sprintf("a_test_job_long_name%v", i),
			Color: color,
		})
	}
	return resultFromJenkins, nil
}

func (api *MockApi) GetCurrentStatus(job string) (status *JobStatus, err error) {
	var culprits []Culprit = make([]Culprit, 0)
	for i := 0; i < rand.Intn(5); i++ {
		culprits = append(culprits, Culprit{
			FullName: random_causes[rand.Intn(len(random_causes))],
		})
	}
	var causes []Cause = make([]Cause, 0)
	for i := 0; i < rand.Intn(5); i++ {
		causes = append(causes, Cause{
			//TODO: mock also causes chain here
			UserId: random_causes[rand.Intn(len(random_causes))],
		})
	}
	result := &JobStatus{
		Building: rand.Intn(2) == 0,
		EstimatedDuration: int64(rand.Intn(300000)),
		Timestamp: time.Now().UnixNano() / 1000 / 1000 - int64(rand.Intn(300000)),
		Culprits: culprits,
		Actions:[]Action{
			Action {
				Causes: causes,
			},
		},
	}
	return result, nil
}

func (api *MockApi) CausesFriendly(status *JobStatus) string {
	var result []string = make([]string, 0)
	var random_causes = []string{
		"milan",
		"fred",
		"johnny",
		"unknown",
	}
	for i := 0; i < rand.Intn(5); i++ {
		result = append(result, random_causes[rand.Intn(len(random_causes))])
	}
	return strings.Join(result, ", ")
}

func (api *MockApi) GetLastBuildUrlForJob(job string) string {
	return fmt.Sprintf("http://mock_jenkins/job/%v/lastBuild/", job)
}
