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
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type Build struct {
	Raw     *BuildResponse
	Job     *Job
	Jenkins *Jenkins
	Base    string
	Depth   int
}

type parameter struct {
	Name  string
	Value string
}

type branch struct {
	SHA1 string
	Name string
}

type BuildRevision struct {
	SHA1   string   `json:"SHA1"`
	Branch []branch `json:"branch"`
}

type Builds struct {
	BuildNumber int64         `json:"buildNumber"`
	BuildResult interface{}   `json:"buildResult"`
	Marked      BuildRevision `json:"marked"`
	Revision    BuildRevision `json:"revision"`
}

type Culprit struct {
	AbsoluteUrl string
	FullName    string
}

type generalObj struct {
	Parameters              []parameter              `json:"parameters"`
	Causes                  []map[string]interface{} `json:"causes"`
	BuildsByBranchName      map[string]Builds        `json:"buildsByBranchName"`
	LastBuiltRevision       BuildRevision            `json:"lastBuiltRevision"`
	RemoteUrls              []string                 `json:"remoteUrls"`
	ScmName                 string                   `json:"scmName"`
	MercurialNodeName       string                   `json:"mercurialNodeName"`
	MercurialRevisionNumber string                   `json:"mercurialRevisionNumber"`
	Subdir                  interface{}              `json:"subdir"`
	TotalCount              int64
	UrlName                 string
}

type TestResult struct {
	Duration  float64 `json:"duration"`
	Empty     bool    `json:"empty"`
	FailCount int64   `json:"failCount"`
	PassCount int64   `json:"passCount"`
	SkipCount int64   `json:"skipCount"`
	Suites    []struct {
		Cases []struct {
			Age             int64       `json:"age"`
			ClassName       string      `json:"className"`
			Duration        float64     `json:"duration"`
			ErrorDetails    interface{} `json:"errorDetails"`
			ErrorStackTrace interface{} `json:"errorStackTrace"`
			FailedSince     int64       `json:"failedSince"`
			Name            string      `json:"name"`
			Skipped         bool        `json:"skipped"`
			SkippedMessage  interface{} `json:"skippedMessage"`
			Status          string      `json:"status"`
			Stderr          interface{} `json:"stderr"`
			Stdout          interface{} `json:"stdout"`
		} `json:"cases"`
		Duration  float64     `json:"duration"`
		ID        interface{} `json:"id"`
		Name      string      `json:"name"`
		Stderr    interface{} `json:"stderr"`
		Stdout    interface{} `json:"stdout"`
		Timestamp interface{} `json:"timestamp"`
	} `json:"suites"`
}

type BuildResponse struct {
	Actions   []generalObj
	Artifacts []struct {
		DisplayPath  string `json:"displayPath"`
		FileName     string `json:"fileName"`
		RelativePath string `json:"relativePath"`
	} `json:"artifacts"`
	Building  bool   `json:"building"`
	BuiltOn   string `json:"builtOn"`
	ChangeSet struct {
		Items []struct {
			AffectedPaths []string `json:"affectedPaths"`
			Author        struct {
				AbsoluteUrl string `json:"absoluteUrl"`
				FullName    string `json:"fullName"`
			} `json:"author"`
			Comment  string `json:"comment"`
			CommitID string `json:"commitId"`
			Date     string `json:"date"`
			ID       string `json:"id"`
			Msg      string `json:"msg"`
			Paths    []struct {
				EditType string `json:"editType"`
				File     string `json:"file"`
			} `json:"paths"`
			Timestamp int64 `json:"timestamp"`
		} `json:"items"`
		Kind      string `json:"kind"`
		Revisions []struct {
			Module   string
			Revision int
		} `json:"revision"`
	} `json:"changeSet"`
	ChangeSets []struct {
		Items []struct {
			AffectedPaths []string `json:"affectedPaths"`
			Author        struct {
				AbsoluteUrl string `json:"absoluteUrl"`
				FullName    string `json:"fullName"`
			} `json:"author"`
			Comment  string `json:"comment"`
			CommitID string `json:"commitId"`
			Date     string `json:"date"`
			ID       string `json:"id"`
			Msg      string `json:"msg"`
			Paths    []struct {
				EditType string `json:"editType"`
				File     string `json:"file"`
			} `json:"paths"`
			Timestamp int64 `json:"timestamp"`
		} `json:"items"`
		Kind      string `json:"kind"`
		Revisions []struct {
			Module   string
			Revision int
		} `json:"revision"`
	} `json:"changeSets"`
	Culprits          []Culprit   `json:"culprits"`
	Description       interface{} `json:"description"`
	Duration          float64     `json:"duration"`
	EstimatedDuration float64     `json:"estimatedDuration"`
	Executor          interface{} `json:"executor"`
	DisplayName       string      `json:"displayName"`
	FullDisplayName   string      `json:"fullDisplayName"`
	ID                string      `json:"id"`
	KeepLog           bool        `json:"keepLog"`
	Number            int64       `json:"number"`
	QueueID           int64       `json:"queueId"`
	Result            string      `json:"result"`
	Timestamp         int64       `json:"timestamp"`
	URL               string      `json:"url"`
	MavenArtifacts    interface{} `json:"mavenArtifacts"`
	MavenVersionUsed  string      `json:"mavenVersionUsed"`
	FingerPrint       []FingerPrintResponse
	Runs              []struct {
		Number int64
		URL    string
	} `json:"runs"`
}

type consoleResponse struct {
	Content     string
	Offset      int64
	HasMoreText bool
}

// Builds
func (b *Build) Info() *BuildResponse {
	return b.Raw
}

func (b *Build) GetActions() []generalObj {
	return b.Raw.Actions
}

func (b *Build) GetUrl() string {
	return b.Raw.URL
}

func (b *Build) GetBuildNumber() int64 {
	return b.Raw.Number
}
func (b *Build) GetResult() string {
	return b.Raw.Result
}

func (b *Build) GetArtifacts() []Artifact {
	artifacts := make([]Artifact, len(b.Raw.Artifacts))
	for i, artifact := range b.Raw.Artifacts {
		artifacts[i] = Artifact{
			Jenkins:  b.Jenkins,
			Build:    b,
			FileName: artifact.FileName,
			Path:     b.Base + "/artifact/" + artifact.RelativePath,
		}
	}
	return artifacts
}

func (b *Build) GetCulprits() []Culprit {
	return b.Raw.Culprits
}

func (b *Build) Stop(ctx context.Context) (bool, error) {
	if b.IsRunning(ctx) {
		response, err := b.Jenkins.Requester.Post(ctx, b.Base+"/stop", nil, nil, nil)
		if err != nil {
			return false, err
		}
		return response.StatusCode == 200, nil
	}
	return true, nil
}

func (b *Build) Term(ctx context.Context) (bool, error) {
	if b.IsRunning(ctx) {
		response, err := b.Jenkins.Requester.Post(ctx, b.Base+"/term", nil, nil, nil)
		if err != nil {
			return false, err
		}
		return response.StatusCode == 200, nil
	}
	return true, nil
}

func (b *Build) Kill(ctx context.Context) (bool, error) {
	if b.IsRunning(ctx) {
		response, err := b.Jenkins.Requester.Post(ctx, b.Base+"/kill", nil, nil, nil)
		if err != nil {
			return false, err
		}
		return response.StatusCode == 200, nil
	}
	return true, nil
}

func (b *Build) GetConsoleOutput(ctx context.Context) string {
	url := b.Base + "/consoleText"
	var content string
	b.Jenkins.Requester.GetXML(ctx, url, &content, nil)
	return content
}

func (b *Build) GetConsoleOutputFromIndex(ctx context.Context, startID int64) (consoleResponse, error) {
	strstart := strconv.FormatInt(startID, 10)
	url := b.Base + "/logText/progressiveText"

	var console consoleResponse

	querymap := make(map[string]string)
	querymap["start"] = strstart
	rsp, err := b.Jenkins.Requester.Get(ctx, url, &console.Content, querymap)
	if err != nil {
		return console, err
	}

	textSize := rsp.Header.Get("X-Text-Size")
	console.HasMoreText = len(rsp.Header.Get("X-More-Data")) != 0
	console.Offset, err = strconv.ParseInt(textSize, 10, 64)
	if err != nil {
		return console, err
	}

	return console, err
}

func (b *Build) GetCauses(ctx context.Context) ([]map[string]interface{}, error) {
	_, err := b.Poll(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range b.Raw.Actions {
		if a.Causes != nil {
			return a.Causes, nil
		}
	}
	return nil, errors.New("No Causes")
}

func (b *Build) GetParameters() []parameter {
	for _, a := range b.Raw.Actions {
		if a.Parameters != nil {
			return a.Parameters
		}
	}
	return nil
}

func (b *Build) GetInjectedEnvVars(ctx context.Context) (map[string]string, error) {
	var envVars struct {
		EnvMap map[string]string `json:"envMap"`
	}
	endpoint := b.Base + "/injectedEnvVars"
	_, err := b.Jenkins.Requester.GetJSON(ctx, endpoint, &envVars, nil)
	if err != nil {
		return envVars.EnvMap, err
	}
	return envVars.EnvMap, nil
}

func (b *Build) GetDownstreamBuilds(ctx context.Context) ([]*Build, error) {
	result := make([]*Build, 0)
	downstreamJobs, err := b.Job.GetDownstreamJobs(ctx)
	if err != nil {
		return nil, err
	}
	for _, job := range downstreamJobs {
		allBuildIDs, err := job.GetAllBuildIds(ctx)
		if err != nil {
			return nil, err
		}
		for _, buildID := range allBuildIDs {
			build, err := job.GetBuild(ctx, buildID.Number)
			if err != nil {
				return nil, err
			}
			upstreamBuild, err := build.GetUpstreamBuild(ctx)
			// older build may no longer exist, so simply ignore these
			// cannot compare only id, it can be from different job
			if err == nil && b.GetUrl() == upstreamBuild.GetUrl() {
				result = append(result, build)
				break
			}
		}
	}
	return result, nil
}

func (b *Build) GetDownstreamJobNames(ctx context.Context) []string {
	result := make([]string, 0)
	downstreamJobs := b.Job.GetDownstreamJobsMetadata()
	fingerprints := b.GetAllFingerPrints(ctx)
	for _, fingerprint := range fingerprints {
		for _, usage := range fingerprint.Raw.Usage {
			for _, job := range downstreamJobs {
				if job.Name == usage.Name {
					result = append(result, job.Name)
				}
			}
		}
	}
	return result
}

func (b *Build) GetAllFingerPrints(ctx context.Context) []*FingerPrint {
	b.Poll(ctx)
	result := make([]*FingerPrint, len(b.Raw.FingerPrint))
	for i, f := range b.Raw.FingerPrint {
		result[i] = &FingerPrint{Jenkins: b.Jenkins, Base: "/fingerprint/", Id: f.Hash, Raw: &f}
	}
	return result
}

func (b *Build) GetUpstreamJob(ctx context.Context) (*Job, error) {
	causes, err := b.GetCauses(ctx)
	if err != nil {
		return nil, err
	}

	for _, cause := range causes {
		if job, ok := cause["upstreamProject"]; ok {
			return b.Jenkins.GetJob(ctx, job.(string))
		}
	}
	return nil, errors.New("Unable to get Upstream Job")
}

func (b *Build) GetUpstreamBuildNumber(ctx context.Context) (int64, error) {
	causes, err := b.GetCauses(ctx)
	if err != nil {
		return 0, err
	}
	for _, cause := range causes {
		if build, ok := cause["upstreamBuild"]; ok {
			switch t := build.(type) {
			default:
				return t.(int64), nil
			case float64:
				return int64(t), nil
			}
		}
	}
	return 0, nil
}

func (b *Build) GetUpstreamBuild(ctx context.Context) (*Build, error) {
	job, err := b.GetUpstreamJob(ctx)
	if err != nil {
		return nil, err
	}
	if job != nil {
		buildNumber, err := b.GetUpstreamBuildNumber(ctx)
		if err == nil && buildNumber != 0 {
			return job.GetBuild(ctx, buildNumber)
		}
	}
	return nil, errors.New("Build not found")
}

func (b *Build) GetMatrixRuns(ctx context.Context) ([]*Build, error) {
	_, err := b.Poll(ctx, 0)
	if err != nil {
		return nil, err
	}
	runs := b.Raw.Runs
	result := make([]*Build, len(b.Raw.Runs))
	r, _ := regexp.Compile(`job/(.*?)/(.*?)/(\d+)/`)

	for i, run := range runs {
		result[i] = &Build{Jenkins: b.Jenkins, Job: b.Job, Raw: new(BuildResponse), Depth: 1, Base: "/" + r.FindString(run.URL)}
		result[i].Poll(ctx)
	}
	return result, nil
}

func (b *Build) GetResultSet(ctx context.Context) (*TestResult, error) {

	url := b.Base + "/testReport"
	var report TestResult

	_, err := b.Jenkins.Requester.GetJSON(ctx, url, &report, nil)
	if err != nil {
		return nil, err
	}

	return &report, nil

}

func (b *Build) GetTimestamp() time.Time {
	msInt := int64(b.Raw.Timestamp)
	return time.Unix(0, msInt*int64(time.Millisecond))
}

func (b *Build) GetDuration() float64 {
	return b.Raw.Duration
}

func (b *Build) GetRevision() string {
	vcs := b.Raw.ChangeSet.Kind

	if vcs == "git" || vcs == "hg" {
		for _, a := range b.Raw.Actions {
			if a.LastBuiltRevision.SHA1 != "" {
				return a.LastBuiltRevision.SHA1
			}
			if a.MercurialRevisionNumber != "" {
				return a.MercurialRevisionNumber
			}
		}
	} else if vcs == "svn" {
		return strconv.Itoa(b.Raw.ChangeSet.Revisions[0].Revision)
	}
	return ""
}

func (b *Build) GetRevisionBranch() string {
	vcs := b.Raw.ChangeSet.Kind
	if vcs == "git" {
		for _, a := range b.Raw.Actions {
			if len(a.LastBuiltRevision.Branch) > 0 && a.LastBuiltRevision.Branch[0].SHA1 != "" {
				return a.LastBuiltRevision.Branch[0].SHA1
			}
		}
	} else {
		panic("Not implemented")
	}
	return ""
}

func (b *Build) IsGood(ctx context.Context) bool {
	return (!b.IsRunning(ctx) && b.Raw.Result == STATUS_SUCCESS)
}

func (b *Build) IsRunning(ctx context.Context) bool {
	_, err := b.Poll(ctx)
	if err != nil {
		return false
	}
	return b.Raw.Building
}

func (b *Build) SetDescription(ctx context.Context, description string) error {
	data := url.Values{}
	data.Set("description", description)
	_, err := b.Jenkins.Requester.Post(ctx, b.Base+"/submitDescription", bytes.NewBufferString(data.Encode()), nil, nil)
	return err
}

// Poll for current data. Optional parameter - depth.
// More about depth here: https://wiki.jenkins-ci.org/display/JENKINS/Remote+access+API
func (b *Build) Poll(ctx context.Context, options ...interface{}) (int, error) {
	depth := "-1"

	for _, o := range options {
		switch v := o.(type) {
		case string:
			depth = v
		case int:
			depth = strconv.Itoa(v)
		case int64:
			depth = strconv.FormatInt(v, 10)
		}
	}
	if depth == "-1" {
		depth = strconv.Itoa(b.Depth)
	}

	qr := map[string]string{
		"depth": depth,
	}
	response, err := b.Jenkins.Requester.GetJSON(ctx, b.Base, b.Raw, qr)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
