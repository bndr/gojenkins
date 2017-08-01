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
	"errors"
	"strconv"
)

// View represents a view
type View struct {
	Raw     *ViewResponse
	Jenkins *Jenkins
	Base    string
}

// ViewResponse represents a View JSON response
type ViewResponse struct {
	Description string        `json:"description"`
	Jobs        []InnerJob    `json:"jobs"`
	Name        string        `json:"name"`
	Property    []interface{} `json:"property"`
	URL         string        `json:"url"`
}

var (
	// LIST_VIEW is the string matching list view
	LIST_VIEW = "hudson.model.ListView"
	// NESTED_VIEW is the string matching nested view
	NESTED_VIEW = "hudson.plugins.nested_view.NestedView"
	// MY_VIEW is the string matching a personal view
	MY_VIEW = "hudson.model.MyView"
	// DASHBOARD_VIEW is the string matching a dashboard view
	DASHBOARD_VIEW = "hudson.plugins.view.dashboard.Dashboard"
	// PIPELINE_VIEW is the string matching a pipeline view
	PIPELINE_VIEW = "au.com.centrumsystems.hudson.plugin.buildpipeline.BuildPipelineView"
)

// AddJob adds a job. Returns True if successfully added Job, otherwise false
func (v *View) AddJob(name string) (bool, error) {
	url := "/addJobToView"
	qr := map[string]string{"name": name}
	resp, err := v.Jenkins.Requester.Post(v.Base+url, nil, nil, qr)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	}
	return false, errors.New(strconv.Itoa(resp.StatusCode))
}

// DeleteJob deletes a job. Returns True if successfully deleted Job, otherwise false
func (v *View) DeleteJob(name string) (bool, error) {
	url := "/removeJobFromView"
	qr := map[string]string{"name": name}
	resp, err := v.Jenkins.Requester.Post(v.Base+url, nil, nil, qr)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == 200 {
		return true, nil
	}
	return false, errors.New(strconv.Itoa(resp.StatusCode))
}

// GetDescription gets a view's description
func (v *View) GetDescription() string {
	return v.Raw.Description
}

// GetJobs gets a view's jobs
func (v *View) GetJobs() []InnerJob {
	return v.Raw.Jobs
}

// GetName gets a view's name
func (v *View) GetName() string {
	return v.Raw.Name
}

//GetUrl gets a view's url
func (v *View) GetUrl() string {
	return v.Raw.URL
}

// Poll polls a view
func (v *View) Poll() (int, error) {
	response, err := v.Jenkins.Requester.GetJSON(v.Base, v.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
