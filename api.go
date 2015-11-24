package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
"log"
)

type Api interface {
	GetRunningJobs() (resultFromJenkins *JenkinsStatus, err error)
	GetCurrentStatus(job string) (status *JobStatus, err error)
	CausesFriendly(status *JobStatus) string
	GetLastBuildUrlForJob(job string) string
}

type JenkinsApi struct {
	ServerLocation string
	cachedCauses   map[string]([]string)
}

// DTOs
// 1. DON'T TRY to use camelCase in DTOs, json unmarshaller doesn't see it!
// 2. DON'T TRY to put space between ":" and "\"", unmarshaller doesn't see it (sometimes)!

type JobStatus struct {
	Building          bool      `json:"building"`
	Actions           []Action  `json:"actions"`
	Culprits          []Culprit `json:"culprits"`
	EstimatedDuration int64     `json:"estimatedDuration"`
	Timestamp         int64     `json:"timestamp"`
}

func (status *JobStatus) CulpritsFriendly() string {
	var result []string = make([]string, 0)
	for _, culprit := range status.Culprits {
		result = append(result, culprit.FullName)
	}
	return strings.Join(result, ", ")
}

type Culprit struct {
	FullName string `json:"fullName"`
}

type Action struct {
	Causes []Cause `json:"causes"`
}

type Cause struct {
	UserId          string `json:"userId"`
	UpstreamBuild   int `json:"upstreamBuild"`
	UpstreamProject string `json:"upstreamProject"`
	ShortDescription string `json:"shortDescription"`
}

type JenkinsStatus struct {
	JobBuildStatus []JobBuildStatus `json:"jobs"`
}

type JobBuildStatus struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (api *JenkinsApi) GetCurrentStatus(job string) (status *JobStatus, err error) {
	resp, err := http.Get(fmt.Sprintf("%v/job/%v/lastBuild/api/json?pretty=true&tree=timestamp,estimatedDuration,building,culprits[fullName],actions[causes[userId,upstreamBuild,upstreamProject]]", api.ServerLocation, job))
	defer resp.Body.Close()
	if err != nil {
		return
	}
	result := &JobStatus{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (api *JenkinsApi) GetRunningJobs() (resultFromJenkins *JenkinsStatus, err error) {
	resp, err := http.Get(fmt.Sprintf("%v/api/json?pretty=true&tree=jobs[name,color]", api.ServerLocation))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	resultFromJenkins = &JenkinsStatus{}
	err = json.NewDecoder(resp.Body).Decode(&resultFromJenkins)
	return resultFromJenkins, nil
}

func (api *JenkinsApi) CausesFriendly(status *JobStatus) string {
	set := make(map[string]bool, 0)
	for _, culprit := range status.Culprits {
		set[culprit.FullName] = true
	}
	for _, action := range status.Actions {
		for _, cause := range action.Causes {
			if cause.UserId != "" {
				set[cause.UserId] = true
			} else if cause.UpstreamBuild != 0 && cause.UpstreamProject != "" {
				new, err := api.AddCauses(cause.UpstreamProject, cause.UpstreamBuild)
				if err != nil {
					log.Println("Could not catch causes: %v", err);
				} else {
					for _, new := range new {
						set[new] = true
					}
				}
			}
		}
	}
	return joinKeysInCsv(set)
}

func (api *JenkinsApi) AddCauses(upstreamProject string, upstreamBuild int) (target []string, err error) {
	link := fmt.Sprintf("%v/job/%v/%v/api/json?pretty=true&tree=culprits[fullName],actions[causes[userId,upstreamBuild,upstreamProject,shortDescription]]",
		api.ServerLocation, upstreamProject, upstreamBuild)
	if cachedValue, ok := api.cachedCauses[link]; ok {
		return cachedValue, nil
	}
	resp, err := http.Get(link)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	status := JobStatus{}
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return
	}

	target = make([]string, 0)
	for _, action := range status.Actions {
		for _, cause := range action.Causes {
			if cause.UserId != "" {
				target = append(target, cause.UserId)
			} else if cause.UpstreamBuild != 0 && cause.UpstreamProject != "" {
				new, err2 := api.AddCauses(cause.UpstreamProject, cause.UpstreamBuild)
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
	log.Printf("link: %v => %v", link, target)
	if api.cachedCauses == nil {
		api.cachedCauses = make(map[string]([]string), 0)
	}
	api.cachedCauses[link] = target
	return
}

func (api *JenkinsApi) GetLastBuildUrlForJob(job string) string {
	return fmt.Sprintf("%v/job/%v/lastBuild/", api.ServerLocation, job)
}