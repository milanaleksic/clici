package main
import (
	"testing"
	"encoding/json"
	"strings"
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
	status := JenkinsStatus{}
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
	if status.Building  {
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
	if !status.Building  {
		t.Fatal("Did not parse Building")
	}
}