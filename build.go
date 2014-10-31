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
	"strconv"
	"time"
)

type Build struct {
	Raw     *buildResponse
	Job     *Job
	Jenkins *Jenkins
	Base    string
	Depth   int
}

type Parameter struct {
	Name  string
	Value string
}

type Branch struct {
	SHA1 string
	Name string
}

type BuildRevision struct {
	SHA1   string   `json:"SHA1"`
	Branch []Branch `json:"branch"`
}

type Builds struct {
	BuildNumber int           `json:"buildNumber"`
	BuildResult interface{}   `json:"buildResult"`
	Marked      BuildRevision `json:"marked"`
	Revision    BuildRevision `json:"revision"`
}

type Culprit struct {
	AbsoluteUrl string
	FullName    string
}

type GeneralObj struct {
	Parameters              []Parameter              `json:"parameters"`
	Causes                  []map[string]interface{} `json:"causes"`
	BuildsByBranchName      map[string]Builds        `json:"buildsByBranchName"`
	LastBuiltRevision       BuildRevision            `json:"lastBuiltRevision"`
	RemoteUrls              []string                 `json:"remoteUrls"`
	ScmName                 string                   `json:"scmName"`
	MercurialNodeName       string                   `json:"mercurialNodeName"`
	MercurialRevisionNumber string                   `json:"mercurialRevisionNumber"`
	Subdir                  interface{}              `json:"subdir"`
	TotalCount              int
	UrlName                 string
}

type TestResult struct {
	Duration  int64 `json:"duration"`
	Empty     bool  `json:"empty"`
	FailCount int64 `json:"failCount"`
	PassCount int64 `json:"passCount"`
	SkipCount int64 `json:"skipCount"`
	Suites    []struct {
		Cases []struct {
			Age             int64       `json:"age"`
			ClassName       string      `json:"className"`
			Duration        int64       `json:"duration"`
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
		Duration  int64       `json:"duration"`
		ID        interface{} `json:"id"`
		Name      string      `json:"name"`
		Stderr    interface{} `json:"stderr"`
		Stdout    interface{} `json:"stdout"`
		Timestamp interface{} `json:"timestamp"`
	} `json:"suites"`
}

type buildResponse struct {
	Actions   []GeneralObj
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
			CommitId string `json:"commitId"`
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
	Culprits          []Culprit   `json:"culprits"`
	Description       interface{} `json:"description"`
	Duration          int         `json:"duration"`
	EstimatedDuration int         `json:"estimatedDuration"`
	Executor          interface{} `json:"executor"`
	FullDisplayName   string      `json:"fullDisplayName"`
	ID                string      `json:"id"`
	KeepLog           bool        `json:"keepLog"`
	Number            int         `json:"number"`
	Result            string      `json:"result"`
	Timestamp         int         `json:"timestamp"`
	URL               string      `json:"url"`
	MavenArtifacts    interface{} `json:"mavenArtifacts"`
	MavenVersionUsed  string      `json:"mavenVersionUsed"`
}

// Builds
func (b *Build) Info() *buildResponse {
	return b.Raw
}

func (b *Build) GetActions() []GeneralObj {
	return b.Raw.Actions
}

func (b *Build) GetUrl() string {
	return b.Raw.URL
}

func (b *Build) GetBuildNumber() int {
	return b.Raw.Number
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

func (b *Build) Stop() bool {
	if b.IsRunning() {
		b.Jenkins.Requester.GetJSON(b.Base+"/stop", nil, nil)
		return b.Jenkins.Requester.LastResponse.StatusCode == 200
	}
	return true
}

func (b *Build) GetConsoleOutput() string {
	url := b.Base + "/consoleText"
	var content string
	b.Jenkins.Requester.GetXML(url, &content, nil)
	return content
}

func (b *Build) GetCauses() []map[string]interface{} {
	b.Poll()
	for _, a := range b.Raw.Actions {
		if a.Causes != nil {
			return a.Causes
		}
	}
	return nil
}

func (b *Build) GetParameters() []Parameter {
	for _, a := range b.Raw.Actions {
		if a.Parameters != nil {
			return a.Parameters
		}
	}
	return nil
}

func (b *Build) GetDownstreamBuilds() {
	panic("Not Implemented")
}

func (b *Build) GetDownstreamJobs() {
	panic("Not Implemented")
}

func (b *Build) GetUpstreamJob() *Job {
	causes := b.GetCauses()
	if len(causes) > 0 {
		if job, ok := causes[0]["upstreamProject"]; ok {
			return b.Jenkins.GetJob(job.(string))
		}
	}
	return nil
}

func (b *Build) GetUpstreamBuildNumber() string {
	causes := b.GetCauses()
	if len(causes) > 0 {
		if build, ok := causes[0]["upstreamBuild"]; ok {
			return build.(string)
		}
	}
	return ""
}

func (b *Build) GetUpstreamBuild() *Build {
	job := b.GetUpstreamJob()
	if job != nil {
		buildNumber := b.GetUpstreamBuildNumber()
		if len(buildNumber) > 0 {
			return job.GetBuild(b.GetUpstreamBuildNumber())
		}
	}
	return nil
}

func (b *Build) GetMatrixRuns() {
	panic("Not Implemented")
}

func (b *Build) GetResultSet() *TestResult {

	for _, a := range b.Raw.Actions {
		if a.TotalCount == 0 && a.UrlName == "" {
			return nil
		}
	}
	url := b.Base + "/testReport"
	var report TestResult
	b.Jenkins.Requester.GetJSON(url, &report, nil)
	if b.Jenkins.Requester.LastResponse.StatusCode == 200 {
		return &report
	} else {
		return nil
	}
}

func (b *Build) GetTimestamp() time.Time {
	msInt := int64(b.Raw.Timestamp)
	return time.Unix(0, msInt*int64(time.Millisecond))
}

func (b *Build) GetDuration() int {
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

func (b *Build) GetRevistionBranch() string {
	vcs := b.Raw.ChangeSet.Kind
	if vcs == "git" {
		for _, a := range b.Raw.Actions {
			if a.LastBuiltRevision.Branch[0].SHA1 != "" {
				return a.LastBuiltRevision.Branch[0].SHA1
			}
		}
	} else {
		panic("Not implemented")
	}
	return ""
}

func (b *Build) IsGood() bool {
	return (!b.IsRunning() && b.Raw.Result == STATUS_SUCCESS)
}

func (b *Build) IsRunning() bool {
	b.Poll()
	return b.Raw.Building
}

func (b *Build) Poll() int {
	qr := map[string]string{
		"depth": strconv.Itoa(b.Depth),
	}
	b.Jenkins.Requester.GetJSON(b.Base, b.Raw, qr)
	return b.Jenkins.Requester.LastResponse.StatusCode
}
