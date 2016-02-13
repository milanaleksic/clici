package jenkins

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParsingBuildStatus(t *testing.T) {
	var statusWire = `{
  "jobs" : [
    {
      "name" : "Adyen_merchant_onboarding",
      "color" : "blue"
    }
  ]
}
`
	status := Status{}
	err := json.NewDecoder(strings.NewReader(statusWire)).Decode(&status)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.JobBuildStatus) != 1 {
		t.Fatal("Did not parse out the job!")
	}
}

func TestParsingStatusFinished(t *testing.T) {
	var statusWire = `{
  "building" : false,
  "culprits" : [
    {
      "fullName":"milan"
    }
  ]
}
`
	status := JobStatus{}
	err := json.Unmarshal([]byte(statusWire), &status)
	if err != nil {
		t.Fatal(err)
	}
	if status.Building {
		t.Fatal("Did not parse result")
	}
	if status.Culprits[0].FullName != "milan" {
		t.Fatal("Did not parse culprit")
	}
}

func TestParsingStatusActive(t *testing.T) {
	var statusWire = `{
  "result" : "SUCCESS",
  "building" : true,
  "estimatedDuration" : 1128776,
  "id" : "11695",
  "keepLog" : false,
  "timestamp" : 1448028238654,
  "culprits" : [

  ]
}
`
	status := JobStatus{}
	err := json.Unmarshal([]byte(statusWire), &status)
	if err != nil {
		t.Fatal(err)
	}
	if status.EstimatedDuration != 1128776 {
		t.Fatal("Did not parse duration")
	}
	if status.Timestamp != 1448028238654 {
		t.Fatal("Did not parse duration")
	}
	if !status.Building {
		t.Fatal("Did not parse Building")
	}
}

func TestParsingTestingReport(t *testing.T) {
	var statusWire = `{
  "suites" : [
    {
      "cases" : [
        {
          "className" : "com.foobar.at.FailingTest",
          "name" : "testMethod1",
          "status" : "PASSED"
        },
        {
          "className" : "com.foobar.at.FailingTest",
          "name" : "testMethod2",
          "status" : "REGRESSION"
        },
        {
          "className" : "com.foobar.at.FailingTest",
          "name" : "testMethod3",
          "status" : "FAILED"
        }
      ]
    },
    {
      "cases" : [
        {
          "className" : "com.foobar.at.FailingTest",
          "name" : "testMethodX",
          "status" : "PASSED"
        }
      ]
    }
  ]
}
`
	status := testCaseResult{}
	err := json.Unmarshal([]byte(statusWire), &status)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Suites) != 2 {
		t.Fatal("Did not parse suites")
	}
	if len(status.Suites[0].Cases) != 3 {
		t.Fatal("Did not parse suite 1")
	}
	if status.Suites[0].Cases[0].ClassName != "com.foobar.at.FailingTest" {
		t.Fatal("Did not parse ClassName")
	}
	if status.Suites[0].Cases[2].Status != "FAILED" {
		t.Fatal("Did not parse ClassName")
	}
	if status.Suites[1].Cases[0].Name != "testMethodX" {
		t.Fatal("Did not parse suite 2 case 1 method name")
	}
}
