package main

type Job struct {
	Raw       *jobResponse
	Requester *Requester
}

type jobResponse struct {
	Actions   []interface{} `json:"actions"`
	Buildable bool          `json:"buildable"`
	Builds    []struct {
		Number float64 `json:"number"`
		URL    string  `json:"url"`
	} `json:"builds"`
	Color              string        `json:"color"`
	ConcurrentBuild    bool          `json:"concurrentBuild"`
	Description        string        `json:"description"`
	DisplayName        string        `json:"displayName"`
	DisplayNameOrNull  interface{}   `json:"displayNameOrNull"`
	DownstreamProjects []interface{} `json:"downstreamProjects"`
	FirstBuild         struct {
		Number float64 `json:"number"`
		URL    string  `json:"url"`
	} `json:"firstBuild"`
	HealthReport []struct {
		Description   string  `json:"description"`
		IconClassName string  `json:"iconClassName"`
		IconUrl       string  `json:"iconUrl"`
		Score         float64 `json:"score"`
	} `json:"healthReport"`
	InQueue          bool `json:"inQueue"`
	KeepDependencies bool `json:"keepDependencies"`
	LastBuild        struct {
		Number float64 `json:"number"`
		URL    string  `json:"url"`
	} `json:"lastBuild"`
	LastCompletedBuild struct {
		Number float64 `json:"number"`
		URL    string  `json:"url"`
	} `json:"lastCompletedBuild"`
	LastFailedBuild interface{} `json:"lastFailedBuild"`
	LastStableBuild struct {
		Number float64 `json:"number"`
		URL    string  `json:"url"`
	} `json:"lastStableBuild"`
	LastSuccessfulBuild struct {
		Number float64 `json:"number"`
		URL    string  `json:"url"`
	} `json:"lastSuccessfulBuild"`
	LastUnstableBuild     interface{}   `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild interface{}   `json:"lastUnsuccessfulBuild"`
	Name                  string        `json:"name"`
	NextBuildNumber       float64       `json:"nextBuildNumber"`
	Property              []interface{} `json:"property"`
	QueueItem             interface{}   `json:"queueItem"`
	Scm                   struct{}      `json:"scm"`
	UpstreamProjects      []interface{} `json:"upstreamProjects"`
	URL                   string        `json:"url"`
}

func (j *Job) GetName() {

}

func (j *Job) GetDescription() {

}

func (j *Job) GetDetails() {

}

func (j *Job) GetBuild() {

}

func (j *Job) GetLastGoodBuild() {

}

func (j *Job) GetFirstBuild() {

}

func (j *Job) GetLastBuild() {

}

func (j *Job) GetLastStableBuild() {

}

func (j *Job) GetLastFailedBuild() {

}

func (j *Job) GetLastCompletedBuild() {

}

func (j *Job) GetAllBuilds() {

}

func (j *Job) GetBuildMetaData() {

}

func (j *Job) GetUpstreamJobNames() {

}

func (j *Job) GetDownstreamJobNames() {

}

func (j *Job) GetUpstreamJobs() {

}

func (J *Job) GetDownstreamJobs() {

}

func (j *Job) Enable() {

}

func (j *Job) Disable() {

}

func (j *Job) Delete() {

}

func (j *Job) Rename() {

}

func (j *Job) Exists() {

}
func (j *Job) Create() {

}

func (j *Job) GetConfig() {

}

func (j *Job) SetConfig() {

}

func (j *Job) GetBuildUrl() {

}

func (j *Job) IsQueued() {

}

func (j *Job) IsRunning() {

}

func (j *Job) IsEnabled() {

}

func (j *Job) HasQueuedBuild() {

}

func (j *Job) Invoke() {

}
