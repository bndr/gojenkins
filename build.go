package main

type Build struct {
	Raw       *buildResponse
	Requester *Requester
	Job       *Job
}

type Parameter struct {
	Name  string
	Value string
}

type buildResponse struct {
	Actions   []interface{}
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

func (b *Build) pull() {

}

// Builds
func (b *Build) GetRaw() {

}

func (b *Build) GetActions() {

}

func (b *Build) GetBuiltOn() {

}

func (b *Build) GetDescription() {

}

func (b *Build) GetDuration() {

}

func (b *Build) GetExecutor() {

}

func (b *Build) GetResult() {

}

func (b *Build) GetUrl() {

}

func (b *Build) GetArtifacts() {

}

func (b *Build) GetCulprits() {

}

func (b *Build) Stop() {

}

func (b *Build) GetConsoleOutput() {

}

func (b *Build) GetCauses() []Cause {
	b.pull()
	if len(b.Raw.Actions) > 0 {
		if causes, ok := b.Raw.Actions[0].([]Cause); ok {
			return causes
		}
	}
	return nil
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
