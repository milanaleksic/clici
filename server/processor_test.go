package server

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/milanaleksic/clici/jenkins"
	"log"
)

const username = "Some user"

type testAPI struct {
	color string
}

func (api *testAPI) GetKnownJobs() (resultFromJenkins *jenkins.Status, err error) {
	resultFromJenkins = &jenkins.Status{
		JobBuildStatus: []jenkins.JobBuildStatus{
			jenkins.JobBuildStatus{
				Name:  fmt.Sprintf("job1"),
				Color: api.color,
			},
		},
	}
	return
}

func (api *testAPI) GetCurrentStatus(job string) (status *jenkins.JobStatus, err error) {
	var culprits = make([]jenkins.Culprit, 0)
	for i := 0; i < rand.Intn(5); i++ {
		culprits = append(culprits, jenkins.Culprit{
			FullName: username,
		})
	}
	var causes = make([]jenkins.Cause, 0)
	for i := 0; i < rand.Intn(5); i++ {
		causes = append(causes, jenkins.Cause{
			UserID: username,
		})
	}
	result := &jenkins.JobStatus{
		Building:          rand.Intn(2) == 0,
		EstimatedDuration: int64(rand.Intn(300000)),
		Timestamp:         time.Now().UnixNano()/1000/1000 - int64(rand.Intn(300000)),
		Culprits:          culprits,
		Actions: []jenkins.Action{
			jenkins.Action{
				Causes: causes,
			},
		},
	}
	return result, nil
}

func (api *testAPI) CausesFriendly(status *jenkins.JobStatus) string {
	return username
}

func (api *testAPI) CausesOfPreviousFailuresFriendly(job string) string {
	return api.CausesFriendly(&jenkins.JobStatus{})
}

func (api *testAPI) GetLastBuildURLForJob(job string) string {
	return ""
}

func (api *testAPI) GetLastCompletedBuildURLForJob(job string) string {
	return ""
}

func (api *testAPI) GetFailedTestList(job string) (testCaseResult []jenkins.TestCase, err error) {
	var set []jenkins.TestCase
	var randomTests = []string{
		"test1",
		"test2",
		"test3",
		"test4",
	}
	for i := 0; i < rand.Intn(5); i++ {
		aCase := jenkins.TestCase{
			ClassName: randomTests[rand.Intn(len(randomTests))],
			Name:      randomTests[rand.Intn(len(randomTests))],
			Status:    "FAILED",
		}
		set = append(set, aCase)
	}
	return
}


func TestProcessor(t *testing.T) {
	api := testAPI{ color: "blue" }
	processor := NewProcessorWithSupplier(
		func() jenkins.API { return &api },
	)
	processor.mapping.RegisterClient("12345", registration{
		ConnectionID:   "12345",
		ServerLocation: "localhost",
		JobName:        "job1",
	})
	connIDToJobStates := processor.ProcessMappings()
	if len(connIDToJobStates) != 1 {
		t.Fatalf("No unique mapping change detected, %v", connIDToJobStates)
	}
	if len(connIDToJobStates["12345"]) != 1 {
		t.Fatalf("No unique mapping change detected for id, %v", connIDToJobStates)
	}
	log.Printf("Known job states: %v", connIDToJobStates["12345"][0])
	processor.mapping.UnRegisterClient("12345")

	connIDToJobStates = processor.ProcessMappings()
	if len(connIDToJobStates) != 0 {
		t.Errorf("After de-registration, no registration expected! %v", connIDToJobStates)
	}
}
