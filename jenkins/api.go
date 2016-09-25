package jenkins

// API is defining known and supported calls towards a Jenkins server
type API interface {
	GetKnownJobs() (resultFromJenkins *Status, err error)
	GetCurrentStatus(job string) (status *JobStatus, err error)
	GetStatusForJob(job string, jobID string) (status *JobStatus, err error)
	CausesFriendly(status *JobStatus) string
	CausesOfFailuresFriendly(name, id string) string
	CausesOfPreviousFailuresFriendly(job string) string
	GetLastBuildURLForJob(job string) string
	GetLastCompletedBuildURLForJob(job string) string
	GetFailedTestList(job string) (testCaseResult []TestCase, err error)
	GetFailedTestListFor(job, id string) (testCaseResult []TestCase, err error)
	GetLastLogLines(job, id string, lineCount int) ([]string, error)
}

// NewMockAPI creates mocking API, usable for testing only
func NewMockAPI() API {
	return &MockAPI{}
}

// NewAPI will create a real API, which will communicate with a certain Jenkins server
func NewAPI(location string) API {
	return &ServerAPI{
		ServerLocation: location,
	}
}

// Status represents API response for list of currently known jobs in the Jenkins Server.
type Status struct {
	JobBuildStatus []JobBuildStatus `json:"jobs"`
}

// JobBuildStatus status for a single job
type JobBuildStatus struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// JobStatus contains a parsed Jenkins server response about a single job result status
type JobStatus struct {
	ID                string      `json:"id"`
	Result            string      `json:"result"`
	Building          bool        `json:"building"`
	Actions           []Action    `json:"actions"`
	Culprits          []Culprit   `json:"culprits"`
	ChangeSets        []ChangeSet `json:"changeSets"`
	EstimatedDuration int64       `json:"estimatedDuration"`
	Timestamp         int64       `json:"timestamp"`
}

// ChangeSet carries all changesets included in this run
type ChangeSet struct {
	Items []ChangeSetItem `json:"items"`
}

// ChangeSetItem contains individual change included in this run
type ChangeSetItem struct {
	Author Culprit `json:"author"`
}

// Culprit is a wrapper around a full name for a culprit
type Culprit struct {
	FullName string `json:"fullName"`
}

// Action is a wrapper around causes
type Action struct {
	Causes []Cause `json:"causes"`
}

// Cause is defining a cause for a job execution
type Cause struct {
	UserID           string `json:"userId"`
	UpstreamBuild    int    `json:"upstreamBuild"`
	UpstreamProject  string `json:"upstreamProject"`
	ShortDescription string `json:"shortDescription"`
}

// TestCaseResult is a result of a single test suite execution
type TestCaseResult struct {
	Suites []TestSuite `json:"suites"`
}

// TestSuite is a wrapper around multiple test cases executed in a Jenkins job
type TestSuite struct {
	Cases []TestCase `json:"cases"`
}

// TestCase depicts part of Jenkins API and identifies which particular test case failed while running a job
type TestCase struct {
	ClassName       string `json:"className"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	ErrorStackTrace string `json:"errorStackTrace"`
}
