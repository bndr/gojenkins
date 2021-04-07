// Copyright 2015 Vadim Kravcenko
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type Job struct {
	Raw     *JobResponse
	Jenkins *Jenkins
	Base    string
}

type JobBuild struct {
	Number int64
	URL    string
}

type InnerJob struct {
	Class string `json:"_class"`
	Name  string `json:"name"`
	Url   string `json:"url"`
	Color string `json:"color"`
}

type ParameterDefinition struct {
	DefaultParameterValue struct {
		Name  string      `json:"name"`
		Value interface{} `json:"value"`
	} `json:"defaultParameterValue"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type JobResponse struct {
	Class              string `json:"_class"`
	Actions            []generalObj
	Buildable          bool `json:"buildable"`
	Builds             []JobBuild
	Color              string      `json:"color"`
	ConcurrentBuild    bool        `json:"concurrentBuild"`
	Description        string      `json:"description"`
	DisplayName        string      `json:"displayName"`
	DisplayNameOrNull  interface{} `json:"displayNameOrNull"`
	DownstreamProjects []InnerJob  `json:"downstreamProjects"`
	FirstBuild         JobBuild
	FullName           string `json:"fullName"`
	FullDisplayName    string `json:"fullDisplayName"`
	HealthReport       []struct {
		Description   string `json:"description"`
		IconClassName string `json:"iconClassName"`
		IconUrl       string `json:"iconUrl"`
		Score         int64  `json:"score"`
	} `json:"healthReport"`
	InQueue               bool     `json:"inQueue"`
	KeepDependencies      bool     `json:"keepDependencies"`
	LastBuild             JobBuild `json:"lastBuild"`
	LastCompletedBuild    JobBuild `json:"lastCompletedBuild"`
	LastFailedBuild       JobBuild `json:"lastFailedBuild"`
	LastStableBuild       JobBuild `json:"lastStableBuild"`
	LastSuccessfulBuild   JobBuild `json:"lastSuccessfulBuild"`
	LastUnstableBuild     JobBuild `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild JobBuild `json:"lastUnsuccessfulBuild"`
	Name                  string   `json:"name"`
	NextBuildNumber       int64    `json:"nextBuildNumber"`
	Property              []struct {
		ParameterDefinitions []ParameterDefinition `json:"parameterDefinitions"`
	} `json:"property"`
	QueueItem        interface{} `json:"queueItem"`
	Scm              struct{}    `json:"scm"`
	UpstreamProjects []InnerJob  `json:"upstreamProjects"`
	URL              string      `json:"url"`
	Jobs             []InnerJob  `json:"jobs"`
	PrimaryView      *ViewData   `json:"primaryView"`
	Views            []ViewData  `json:"views"`
}

func (j *Job) parentBase() string {
	return j.Base[:strings.LastIndex(j.Base, "/job/")]
}

type History struct {
	BuildDisplayName string
	BuildNumber      int
	BuildStatus      string
	BuildTimestamp   int64
}

func (j *Job) GetName() string {
	return j.Raw.Name
}

func (j *Job) GetDescription() string {
	return j.Raw.Description
}

func (j *Job) GetDetails() *JobResponse {
	return j.Raw
}

func (j *Job) GetBuild(ctx context.Context, id int64) (*Build, error) {

	// Support customized server URL,
	// i.e. Server : https://<domain>/jenkins/job/JOB1
	// "https://<domain>/jenkins/" is the server URL,
	// we are expecting jobURL = "job/JOB1"
	jobURL := strings.Replace(j.Raw.URL, j.Jenkins.Server, "", -1)
	build := Build{Jenkins: j.Jenkins, Job: j, Raw: new(BuildResponse), Depth: 1, Base: jobURL + "/" + strconv.FormatInt(id, 10)}
	status, err := build.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &build, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Job) getBuildByType(ctx context.Context, buildType string) (*Build, error) {
	allowed := map[string]JobBuild{
		"lastStableBuild":     j.Raw.LastStableBuild,
		"lastSuccessfulBuild": j.Raw.LastSuccessfulBuild,
		"lastBuild":           j.Raw.LastBuild,
		"lastCompletedBuild":  j.Raw.LastCompletedBuild,
		"firstBuild":          j.Raw.FirstBuild,
		"lastFailedBuild":     j.Raw.LastFailedBuild,
	}
	number := ""
	if val, ok := allowed[buildType]; ok {
		number = strconv.FormatInt(val.Number, 10)
	} else {
		panic("No Such Build")
	}
	build := Build{
		Jenkins: j.Jenkins,
		Depth:   1,
		Job:     j,
		Raw:     new(BuildResponse),
		Base:    j.Base + "/" + number}
	status, err := build.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &build, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Job) GetLastSuccessfulBuild(ctx context.Context) (*Build, error) {
	return j.getBuildByType(ctx, "lastSuccessfulBuild")
}

func (j *Job) GetFirstBuild(ctx context.Context) (*Build, error) {
	return j.getBuildByType(ctx, "firstBuild")
}

func (j *Job) GetLastBuild(ctx context.Context) (*Build, error) {
	return j.getBuildByType(ctx, "lastBuild")
}

func (j *Job) GetLastStableBuild(ctx context.Context) (*Build, error) {
	return j.getBuildByType(ctx, "lastStableBuild")
}

func (j *Job) GetLastFailedBuild(ctx context.Context) (*Build, error) {
	return j.getBuildByType(ctx, "lastFailedBuild")
}

func (j *Job) GetLastCompletedBuild(ctx context.Context) (*Build, error) {
	return j.getBuildByType(ctx, "lastCompletedBuild")
}

func (j *Job) GetBuildsFields(ctx context.Context, fields []string, custom interface{}) error {
	if fields == nil || len(fields) == 0 {
		return fmt.Errorf("one or more field value needs to be specified")
	}
	// limit overhead using builds instead of allBuilds, which returns the last 100 build
	_, err := j.Jenkins.Requester.GetJSON(ctx, j.Base, &custom, map[string]string{"tree": "builds[" + strings.Join(fields, ",") + "]"})
	if err != nil {
		return err
	}
	return nil
}

// Returns All Builds with Number and URL
func (j *Job) GetAllBuildIds(ctx context.Context) ([]JobBuild, error) {
	var buildsResp struct {
		Builds []JobBuild `json:"allBuilds"`
	}
	_, err := j.Jenkins.Requester.GetJSON(ctx, j.Base, &buildsResp, map[string]string{"tree": "allBuilds[number,url]"})
	if err != nil {
		return nil, err
	}
	return buildsResp.Builds, nil
}

func (j *Job) GetUpstreamJobsMetadata() []InnerJob {
	return j.Raw.UpstreamProjects
}

func (j *Job) GetDownstreamJobsMetadata() []InnerJob {
	return j.Raw.DownstreamProjects
}

func (j *Job) GetInnerJobsMetadata() []InnerJob {
	return j.Raw.Jobs
}

func (j *Job) GetUpstreamJobs(ctx context.Context) ([]*Job, error) {
	jobs := make([]*Job, len(j.Raw.UpstreamProjects))
	for i, job := range j.Raw.UpstreamProjects {
		ji, err := j.Jenkins.GetJob(ctx, job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

func (j *Job) GetDownstreamJobs(ctx context.Context) ([]*Job, error) {
	jobs := make([]*Job, len(j.Raw.DownstreamProjects))
	for i, job := range j.Raw.DownstreamProjects {
		ji, err := j.Jenkins.GetJob(ctx, job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

func (j *Job) GetInnerJob(ctx context.Context, id string) (*Job, error) {
	job := Job{Jenkins: j.Jenkins, Raw: new(JobResponse), Base: j.Base + "/job/" + id}
	status, err := job.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Job) GetInnerJobs(ctx context.Context) ([]*Job, error) {
	jobs := make([]*Job, len(j.Raw.Jobs))
	for i, job := range j.Raw.Jobs {
		ji, err := j.GetInnerJob(ctx, job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

func (j *Job) Enable(ctx context.Context) (bool, error) {
	resp, err := j.Jenkins.Requester.Post(ctx, j.Base+"/enable", nil, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode))
	}
	return true, nil
}

func (j *Job) Disable(ctx context.Context) (bool, error) {
	resp, err := j.Jenkins.Requester.Post(ctx, j.Base+"/disable", nil, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode))
	}
	return true, nil
}

func (j *Job) Delete(ctx context.Context) (bool, error) {
	resp, err := j.Jenkins.Requester.Post(ctx, j.Base+"/doDelete", nil, nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode))
	}
	return true, nil
}

func (j *Job) Rename(ctx context.Context, name string) (bool, error) {
	data := url.Values{}
	data.Set("newName", name)
	_, err := j.Jenkins.Requester.Post(ctx, j.Base+"/doRename", bytes.NewBufferString(data.Encode()), nil, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (j *Job) Create(ctx context.Context, config string, qr ...interface{}) (*Job, error) {
	var querystring map[string]string
	if len(qr) > 0 {
		querystring = qr[0].(map[string]string)
	}
	resp, err := j.Jenkins.Requester.PostXML(ctx, j.parentBase()+"/createItem", config, j.Raw, querystring)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		j.Poll(ctx)
		return j, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

func (j *Job) Copy(ctx context.Context, destinationName string) (*Job, error) {
	qr := map[string]string{"name": destinationName, "from": j.GetName(), "mode": "copy"}
	resp, err := j.Jenkins.Requester.Post(ctx, j.parentBase()+"/createItem", nil, nil, qr)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 200 {
		newJob := &Job{Jenkins: j.Jenkins, Raw: new(JobResponse), Base: "/job/" + destinationName}
		_, err := newJob.Poll(ctx)
		if err != nil {
			return nil, err
		}
		return newJob, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

func (j *Job) UpdateConfig(ctx context.Context, config string) error {

	var querystring map[string]string

	resp, err := j.Jenkins.Requester.PostXML(ctx, j.Base+"/config.xml", config, nil, querystring)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		j.Poll(ctx)
		return nil
	}
	return errors.New(strconv.Itoa(resp.StatusCode))

}

func (j *Job) GetConfig(ctx context.Context) (string, error) {
	var data string
	_, err := j.Jenkins.Requester.GetXML(ctx, j.Base+"/config.xml", &data, nil)
	if err != nil {
		return "", err
	}
	return data, nil
}

func (j *Job) GetParameters(ctx context.Context) ([]ParameterDefinition, error) {
	_, err := j.Poll(ctx)
	if err != nil {
		return nil, err
	}
	var parameters []ParameterDefinition
	for _, property := range j.Raw.Property {
		parameters = append(parameters, property.ParameterDefinitions...)
	}
	return parameters, nil
}

func (j *Job) IsQueued(ctx context.Context) (bool, error) {
	if _, err := j.Poll(ctx); err != nil {
		return false, err
	}
	return j.Raw.InQueue, nil
}

func (j *Job) IsRunning(ctx context.Context) (bool, error) {
	if _, err := j.Poll(ctx); err != nil {
		return false, err
	}
	lastBuild, err := j.GetLastBuild(ctx)
	if err != nil {
		return false, err
	}
	return lastBuild.IsRunning(ctx), nil
}

func (j *Job) IsEnabled(ctx context.Context) (bool, error) {
	if _, err := j.Poll(ctx); err != nil {
		return false, err
	}
	return j.Raw.Color != "disabled", nil
}

func (j *Job) HasQueuedBuild() {
	panic("Not Implemented yet")
}

func (j *Job) InvokeSimple(ctx context.Context, params map[string]string) (int64, error) {
	isQueued, err := j.IsQueued(ctx)
	if err != nil {
		return 0, err
	}
	if isQueued {
		Error.Printf("%s is already running", j.GetName())
		return 0, nil
	}

	endpoint := "/build"
	parameters, err := j.GetParameters(ctx)
	if err != nil {
		return 0, err
	}
	if len(parameters) > 0 {
		endpoint = "/buildWithParameters"
	}
	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}
	resp, err := j.Jenkins.Requester.Post(ctx, j.Base+endpoint, bytes.NewBufferString(data.Encode()), nil, nil)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return 0, fmt.Errorf("Could not invoke job %q: %s", j.GetName(), resp.Status)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return 0, errors.New("Don't have key \"Location\" in response of header")
	}

	u, err := url.Parse(location)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(path.Base(u.Path), 10, 64)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func (j *Job) Invoke(ctx context.Context, files []string, skipIfRunning bool, params map[string]string, cause string, securityToken string) (bool, error) {
	isQueued, err := j.IsQueued(ctx)
	if err != nil {
		return false, err
	}
	if isQueued {
		Error.Printf("%s is already running", j.GetName())
		return false, nil
	}
	isRunning, err := j.IsRunning(ctx)
	if err != nil {
		return false, err
	}
	if isRunning && skipIfRunning {
		return false, fmt.Errorf("Will not request new build because %s is already running", j.GetName())
	}

	base := "/build"

	// If parameters are specified - url is /builWithParameters
	if params != nil {
		base = "/buildWithParameters"
	} else {
		params = make(map[string]string)
	}

	// If files are specified - url is /build
	if files != nil {
		base = "/build"
	}
	reqParams := map[string]string{}
	buildParams := map[string]string{}
	if securityToken != "" {
		reqParams["token"] = securityToken
	}

	buildParams["json"] = string(makeJson(params))
	b, _ := json.Marshal(buildParams)
	resp, err := j.Jenkins.Requester.PostFiles(ctx, j.Base+base, bytes.NewBuffer(b), nil, reqParams, files)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		return true, nil
	}
	return false, errors.New(strconv.Itoa(resp.StatusCode))
}

func (j *Job) Poll(ctx context.Context) (int, error) {
	response, err := j.Jenkins.Requester.GetJSON(ctx, j.Base, j.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}

func (j *Job) History(ctx context.Context) ([]*History, error) {
	var s string
	_, err := j.Jenkins.Requester.Get(ctx, j.Base+"/buildHistory/ajax", &s, nil)
	if err != nil {
		return nil, err
	}

	return parseBuildHistory(strings.NewReader(s)), nil
}

func (pr *PipelineRun) ProceedInput(ctx context.Context) (bool, error) {
	actions, _ := pr.GetPendingInputActions(ctx)
	data := url.Values{}
	data.Set("inputId", actions[0].ID)
	params := make(map[string]string)
	data.Set("json", makeJson(params))

	href := pr.Base + "/wfapi/inputSubmit"

	resp, err := pr.Job.Jenkins.Requester.Post(ctx, href, bytes.NewBufferString(data.Encode()), nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode))
	}
	return true, nil
}

func (pr *PipelineRun) AbortInput(ctx context.Context) (bool, error) {
	actions, _ := pr.GetPendingInputActions(ctx)
	data := url.Values{}
	params := make(map[string]string)
	data.Set("json", makeJson(params))

	href := pr.Base + "/input/" + actions[0].ID + "/abort"

	resp, err := pr.Job.Jenkins.Requester.Post(ctx, href, bytes.NewBufferString(data.Encode()), nil, nil)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, errors.New(strconv.Itoa(resp.StatusCode))
	}
	return true, nil
}
