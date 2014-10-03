package main

type Build struct {
	Raw       *BuildResponse
	Requester *Requester
}

type BuildResponse struct {
	Actions []struct {
		Causes []struct {
			ShortDescription string      `json:"shortDescription"`
			UserId           interface{} `json:"userId"`
			UserName         string      `json:"userName"`
		} `json:"causes"`
	} `json:"actions"`
	Artifacts []interface{} `json:"artifacts"`
	Building  bool          `json:"building"`
	BuiltOn   string        `json:"builtOn"`
	ChangeSet struct {
		Items []interface{} `json:"items"`
		Kind  interface{}   `json:"kind"`
	} `json:"changeSet"`
	Culprits          []interface{} `json:"culprits"`
	Description       interface{}   `json:"description"`
	Duration          int           `json:"duration"`
	EstimatedDuration int           `json:"estimatedDuration"`
	Executor          interface{}   `json:"executor"`
	FullDisplayName   string        `json:"fullDisplayName"`
	ID                string        `json:"id"`
	KeepLog           bool          `json:"keepLog"`
	Number            int           `json:"number"`
	Result            string        `json:"result"`
	Timestamp         int           `json:"timestamp"`
	URL               string        `json:"url"`
}

// Builds
func (b *Build) GetDetails() {

}

func (b *Build) Stop() {

}

func (b *Build) GetConsoleOutput() {

}

func (b *Build) GetCauses() {

}

func (b *Build) GetDownstreamBuilds() {

}

func (b *Build) GetDownstreamJobs() {

}

func (b *Build) GetUpstreamJobs() {

}

func (b *Build) GetUpstreamBuilds() {

}

func (b *Build) GetMasterBuildNumber() {

}

func (b *Build) GetMasterBuildJob() {

}

func (b *Build) GetMatrixRuns() {

}

func (b *Build) GetResultUrl() {

}

func (b *Build) GetResultSet() {

}

func (b *Build) GetTimestamp() {

}

func (b *Build) IsGood() {

}

func (b *Build) IsRunning() {

}
