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
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestView_GetName(t *testing.T) {
	view := &View{
		Raw: &ViewResponse{
			Name: "All",
		},
	}

	assert.Equal(t, "All", view.GetName())
}

func TestView_GetDescription(t *testing.T) {
	view := &View{
		Raw: &ViewResponse{
			Description: "This is the main view",
		},
	}

	assert.Equal(t, "This is the main view", view.GetDescription())
}

func TestView_GetUrl(t *testing.T) {
	view := &View{
		Raw: &ViewResponse{
			URL: "http://jenkins/view/All/",
		},
	}

	assert.Equal(t, "http://jenkins/view/All/", view.GetUrl())
}

func TestView_GetJobs(t *testing.T) {
	jobs := []InnerJob{
		{Name: "job-1", Url: "http://jenkins/job/job-1", Color: "blue"},
		{Name: "job-2", Url: "http://jenkins/job/job-2", Color: "red"},
		{Name: "job-3", Url: "http://jenkins/job/job-3", Color: "notbuilt"},
	}
	view := &View{
		Raw: &ViewResponse{
			Jobs: jobs,
		},
	}

	result := view.GetJobs()
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "job-1", result[0].Name)
	assert.Equal(t, "blue", result[0].Color)
	assert.Equal(t, "job-2", result[1].Name)
	assert.Equal(t, "red", result[1].Color)
}

func TestView_GetJobs_Empty(t *testing.T) {
	view := &View{
		Raw: &ViewResponse{
			Jobs: []InnerJob{},
		},
	}

	result := view.GetJobs()
	assert.Equal(t, 0, len(result))
}

func TestView_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	view := &View{
		Jenkins: jenkins,
		Raw:     &ViewResponse{},
		Base:    "/view/All",
	}

	status, err := view.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestView_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	view := &View{
		Jenkins: jenkins,
		Raw:     &ViewResponse{},
		Base:    "/view/All",
	}

	_, err := view.Poll(context.Background())
	assert.Error(t, err)
}

func TestView_AddJob_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	view := &View{
		Jenkins: jenkins,
		Raw:     &ViewResponse{},
		Base:    "/view/MyView",
	}

	success, err := view.AddJob(context.Background(), "new-job")
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/view/MyView/addJobToView", mock.lastEndpoint)
}

func TestView_AddJob_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 404,
	}

	view := &View{
		Jenkins: jenkins,
		Raw:     &ViewResponse{},
		Base:    "/view/MyView",
	}

	success, err := view.AddJob(context.Background(), "nonexistent-job")
	assert.Error(t, err)
	assert.False(t, success)
}

func TestView_DeleteJob_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	view := &View{
		Jenkins: jenkins,
		Raw:     &ViewResponse{},
		Base:    "/view/MyView",
	}

	success, err := view.DeleteJob(context.Background(), "job-to-remove")
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/view/MyView/removeJobFromView", mock.lastEndpoint)
}

func TestView_DeleteJob_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 404,
	}

	view := &View{
		Jenkins: jenkins,
		Raw:     &ViewResponse{},
		Base:    "/view/MyView",
	}

	success, err := view.DeleteJob(context.Background(), "nonexistent-job")
	assert.Error(t, err)
	assert.False(t, success)
}

func TestViewConstants(t *testing.T) {
	assert.Equal(t, "hudson.model.ListView", LIST_VIEW)
	assert.Equal(t, "hudson.plugins.nested_view.NestedView", NESTED_VIEW)
	assert.Equal(t, "hudson.model.MyView", MY_VIEW)
	assert.Equal(t, "hudson.plugins.view.dashboard.Dashboard", DASHBOARD_VIEW)
	assert.Equal(t, "au.com.centrumsystems.hudson.plugin.buildpipeline.BuildPipelineView", PIPELINE_VIEW)
}
