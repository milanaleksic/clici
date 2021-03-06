package server

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"log"
	"sync"

	"github.com/milanaleksic/clici/jenkins"
	"github.com/milanaleksic/clici/model"
)

const username = "Some user"

type testAPI struct {
	color string
}

func (api *testAPI) GetKnownJobs() (resultFromJenkins *jenkins.Status, err error) {
	resultFromJenkins = &jenkins.Status{
		JobBuildStatus: []jenkins.JobBuildStatus{
			jenkins.JobBuildStatus{
				Name:  fmt.Sprint("job1"),
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
		Timestamp:         time.Now().UnixNano() / 1000 / 1000 - int64(rand.Intn(300000)),
		Culprits:          culprits,
		Actions: []jenkins.Action{
			jenkins.Action{
				Causes: causes,
			},
		},
	}
	return result, nil
}

func (api *testAPI) GetStatusForJob(job string, id string) (status *jenkins.JobStatus, err error) {
	return api.GetCurrentStatus(job)
}

func (api *testAPI) CausesOfFailures(name, id string) []string {
	return []string{username}
}

func (api *testAPI) Causes(status *jenkins.JobStatus) []string {
	return []string{username}
}

func (api *testAPI) CausesOfPreviousFailures(job string) []string {
	return api.Causes(&jenkins.JobStatus{})
}

func (api *testAPI) GetLastBuildURLForJob(job string) string {
	return ""
}

func (api *testAPI) GetLastCompletedBuildURLForJob(job string) string {
	return ""
}

func (api *testAPI) GetLastLogLines(job, id string, lineCount int) (response []string, err error) {
	return []string{"line1", "line2"}, nil
}

func (api *testAPI) GetFailedTestListFor(job, id string) (testCaseResult []jenkins.TestCase, err error) {
	return api.GetFailedTestList(job)
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

// RunJob will execute a job (expected - without parameters)
func (api *testAPI) RunJob(job string) error {
	return nil
}

func TestProcessor(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	api := testAPI{color: "blue"}
	outputChannel := make(chan model.JobState)
	processor := NewProcessorWithSupplier(
		func(serverLocation string, username, server string) jenkins.API {
			return &api
		},
	)

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		select {
		case jobState := <-outputChannel:
			log.Printf("jobState=%v", jobState)
			wg.Done()
		case <-ticker.C:
			wg.Done()
			t.Fatal("Timed out waiting for the response from processor")
		}
	}()

	processor.RegisterClient("12345", "localhost", "job1", outputChannel)
	defer processor.mapping.UnRegisterClient("12345")

	processor.ProcessMappings()

	wg.Wait()
}
