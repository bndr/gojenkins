// Copyright 2014 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type Job struct {
	Raw     *jobResponse
	Jenkins *Jenkins
	Base    string
}

type ActionsObject struct {
	FailCount  int64
	SkipCount  int64
	TotalCount int64
	UrlName    string
}

type jobBuild struct {
	Number int
	URL    string
}

type updownProject struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Color string `json:"color"`
}

type parameterDefinition struct {
	DefaultParameterValue struct {
		Name  string `json:"name"`
		Value bool   `json:"value"`
	} `json:"defaultParameterValue"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type jobResponse struct {
	Actions            interface{}
	Buildable          bool `json:"buildable"`
	Builds             []jobBuild
	Color              string          `json:"color"`
	ConcurrentBuild    bool            `json:"concurrentBuild"`
	Description        string          `json:"description"`
	DisplayName        string          `json:"displayName"`
	DisplayNameOrNull  interface{}     `json:"displayNameOrNull"`
	DownstreamProjects []updownProject `json:"downstreamProjects"`
	FirstBuild         jobBuild
	HealthReport       []struct {
		Description   string `json:"description"`
		IconClassName string `json:"iconClassName"`
		IconUrl       string `json:"iconUrl"`
		Score         int64  `json:"score"`
	} `json:"healthReport"`
	InQueue               bool     `json:"inQueue"`
	KeepDependencies      bool     `json:"keepDependencies"`
	LastBuild             jobBuild `json:"lastBuild"`
	LastCompletedBuild    jobBuild `json:"lastCompletedBuild"`
	LastFailedBuild       jobBuild `json:"lastFailedBuild"`
	LastStableBuild       jobBuild `json:"lastStableBuild"`
	LastSuccessfulBuild   jobBuild `json:"lastSuccessfulBuild"`
	LastUnstableBuild     jobBuild `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild jobBuild `json:"lastUnsuccessfulBuild"`
	Name                  string   `json:"name"`
	NextBuildNumber       int64    `json:"nextBuildNumber"`
	Property              []struct {
		ParameterDefinitions []parameterDefinition `json:"parameterDefinitions"`
	} `json:"property"`
	QueueItem        interface{}     `json:"queueItem"`
	Scm              struct{}        `json:"scm"`
	UpstreamProjects []updownProject `json:"upstreamProjects"`
	URL              string          `json:"url"`
}

func (j *Job) GetName() string {
	return j.Raw.Name
}

func (j *Job) GetDescription() string {
	return j.Raw.Description
}

func (j *Job) GetDetails() *jobResponse {
	return j.Raw
}

func (j *Job) GetBuild(id string) *Build {
	build := Build{Jenkins: j.Jenkins, Raw: new(buildResponse), Depth: 1, Base: "/job/" + j.GetName() + "/" + id}
	if build.Poll() == 200 {
		return &build
	}
	return nil
}

func (j *Job) getBuildByType(buildType string) *Build {
	allowed := map[string]jobBuild{
		"lastStableBuild":     j.Raw.LastStableBuild,
		"lastSuccessfulBuild": j.Raw.LastSuccessfulBuild,
		"lastBuild":           j.Raw.LastBuild,
		"lastCompletedBuild":  j.Raw.LastCompletedBuild,
		"firstBuild":          j.Raw.FirstBuild,
		"lastFailedBuild":     j.Raw.LastFailedBuild,
	}
	number := ""
	if val, ok := allowed[buildType]; ok {
		number = strconv.Itoa(val.Number)
	} else {
		panic("No Such Build")
	}
	build := Build{
		Jenkins: j.Jenkins,
		Depth:   1,
		Job:     j,
		Raw:     new(buildResponse),
		Base:    "/job/" + j.GetName() + "/" + number}
	if build.Poll() == 200 {
		return &build
	}
	return nil
}

func (j *Job) GetLastSuccessfulBuild() *Build {
	return j.getBuildByType("lastSuccessfulBuild")
}

func (j *Job) GetFirstBuild() *Build {
	return j.getBuildByType("firstBuild")
}

func (j *Job) GetLastBuild() *Build {
	return j.getBuildByType("lastBuild")
}

func (j *Job) GetLastStableBuild() *Build {
	return j.getBuildByType("lastStableBuild")
}

func (j *Job) GetLastFailedBuild() *Build {
	return j.getBuildByType("lastFailedBuild")
}

func (j *Job) GetLastCompletedBuild() *Build {
	return j.getBuildByType("lastCompletedBuild")
}

func (j *Job) GetAllBuilds() {
	j.Poll()
	builds := make([]*Build, len(j.Raw.Builds))
	for i, b := range j.Raw.Builds {
		builds[i] = &Build{
			Jenkins: j.Jenkins,
			Depth:   1,
			Raw:     &buildResponse{Number: b.Number, URL: b.URL},
			Base:    "/job/" + j.GetName() + "/" + string(b.Number)}
	}
}

func (j *Job) GetUpstreamJobsMetadata() []updownProject {
	return j.Raw.UpstreamProjects
}

func (j *Job) GetDownstreamJobsMetadata() []updownProject {
	return j.Raw.DownstreamProjects
}

func (j *Job) GetUpstreamJobs() []*Job {
	jobs := make([]*Job, len(j.Raw.UpstreamProjects))
	for i, job := range j.Raw.UpstreamProjects {
		jobs[i] = &Job{
			Raw: &jobResponse{
				Name:  job.Name,
				Color: job.Color,
				URL:   job.Url},
			Jenkins: j.Jenkins,
			Base:    "/job/" + job.Name,
		}
		jobs[i].Poll()
	}
	return jobs
}

func (j *Job) GetDownstreamJobs() []*Job {
	jobs := make([]*Job, len(j.Raw.DownstreamProjects))
	for i, job := range j.Raw.DownstreamProjects {
		jobs[i] = &Job{
			Raw: &jobResponse{
				Name:  job.Name,
				Color: job.Color,
				URL:   job.Url},
			Jenkins: j.Jenkins,
		}
		jobs[i].Poll()
	}
	return jobs
}

func (j *Job) Enable() bool {
	resp := j.Jenkins.Requester.Post(j.Base+"/enable", nil, nil, nil)
	return resp.StatusCode == 200
}

func (j *Job) Disable() bool {
	resp := j.Jenkins.Requester.Post(j.Base+"/disable", nil, nil, nil)
	return resp.StatusCode == 200
}

func (j *Job) Delete() bool {
	resp := j.Jenkins.Requester.Post(j.Base+"/doDelete", nil, nil, nil)
	return resp.StatusCode == 200
}

func (j *Job) Rename(name string) {
	payload, _ := json.Marshal(map[string]string{"newName": name})
	j.Jenkins.Requester.Post(j.Base+"/doRename", bytes.NewBuffer(payload), nil, nil)
}

func (j *Job) Exists() {

}

func (j *Job) Create(config string) *Job {
	resp := j.Jenkins.Requester.Post("/createItem", bytes.NewBuffer([]byte(config)), j.Raw, nil)
	if resp.Status == "200" {
		return j
	} else {
		return nil
	}
}

func (j *Job) Copy(from string, newName string) *Job {
	qr := map[string]string{"name": newName, "from": from, "mode": "copy"}
	resp := j.Jenkins.Requester.Post("/createItem", nil, nil, qr)
	if resp.StatusCode == 200 {
		return j
	}
	return nil
}

func (j *Job) GetConfig() string {
	var data string
	j.Jenkins.Requester.GetXML(j.Base+"/config.xml", &data, nil)
	return data
}

func (j *Job) GetParameters() []parameterDefinition {
	j.Poll()
	if len(j.Raw.Property) < 1 {
		return nil
	}
	return j.Raw.Property[0].ParameterDefinitions
}

func (j *Job) IsQueued() bool {
	j.Poll()
	return j.Raw.InQueue
}

func (j *Job) IsRunning() {
	j.Poll()
}

func (j *Job) IsEnabled() bool {
	j.Poll()
	return j.Raw.Color != "disabled"
}

func (j *Job) HasQueuedBuild() {

}

func (j *Job) Invoke(files []string, options ...interface{}) bool {
	return true
}

func (j *Job) Poll() int {
	j.Jenkins.Requester.GetJSON(j.Base, j.Raw, nil)
	return j.Jenkins.Requester.LastResponse.StatusCode
}
