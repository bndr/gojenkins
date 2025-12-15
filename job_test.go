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

func TestJob_GetName(t *testing.T) {
	job := &Job{
		Raw: &JobResponse{
			Name: "test-job",
		},
	}

	assert.Equal(t, "test-job", job.GetName())
}

func TestJob_GetDescription(t *testing.T) {
	job := &Job{
		Raw: &JobResponse{
			Description: "This is a test job",
		},
	}

	assert.Equal(t, "This is a test job", job.GetDescription())
}

func TestJob_GetDetails(t *testing.T) {
	rawResponse := &JobResponse{
		Name:        "test-job",
		Description: "Test description",
		Buildable:   true,
	}
	job := &Job{
		Raw: rawResponse,
	}

	assert.Equal(t, rawResponse, job.GetDetails())
}

func TestJob_GetUpstreamJobsMetadata(t *testing.T) {
	upstreamJobs := []InnerJob{
		{Name: "upstream-1", Url: "http://jenkins/job/upstream-1"},
		{Name: "upstream-2", Url: "http://jenkins/job/upstream-2"},
	}
	job := &Job{
		Raw: &JobResponse{
			UpstreamProjects: upstreamJobs,
		},
	}

	result := job.GetUpstreamJobsMetadata()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "upstream-1", result[0].Name)
	assert.Equal(t, "upstream-2", result[1].Name)
}

func TestJob_GetDownstreamJobsMetadata(t *testing.T) {
	downstreamJobs := []InnerJob{
		{Name: "downstream-1", Url: "http://jenkins/job/downstream-1"},
	}
	job := &Job{
		Raw: &JobResponse{
			DownstreamProjects: downstreamJobs,
		},
	}

	result := job.GetDownstreamJobsMetadata()
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "downstream-1", result[0].Name)
}

func TestJob_GetInnerJobsMetadata(t *testing.T) {
	innerJobs := []InnerJob{
		{Name: "inner-1", Url: "http://jenkins/job/folder/job/inner-1"},
		{Name: "inner-2", Url: "http://jenkins/job/folder/job/inner-2"},
		{Name: "inner-3", Url: "http://jenkins/job/folder/job/inner-3"},
	}
	job := &Job{
		Raw: &JobResponse{
			Jobs: innerJobs,
		},
	}

	result := job.GetInnerJobsMetadata()
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "inner-1", result[0].Name)
	assert.Equal(t, "inner-2", result[1].Name)
	assert.Equal(t, "inner-3", result[2].Name)
}

func TestJob_IsEnabled_Enabled(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			Color: "blue",
		},
		Base: "/job/test-job",
	}

	enabled, err := job.IsEnabled(context.Background())
	assert.NoError(t, err)
	assert.True(t, enabled)
}

func TestJob_IsEnabled_Disabled(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			Color: "disabled",
		},
		Base: "/job/test-job",
	}

	enabled, err := job.IsEnabled(context.Background())
	assert.NoError(t, err)
	assert.False(t, enabled)
}

func TestJob_IsQueued_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			InQueue: true,
		},
		Base: "/job/test-job",
	}

	queued, err := job.IsQueued(context.Background())
	assert.NoError(t, err)
	assert.True(t, queued)
}

func TestJob_IsQueued_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			InQueue: false,
		},
		Base: "/job/test-job",
	}

	queued, err := job.IsQueued(context.Background())
	assert.NoError(t, err)
	assert.False(t, queued)
}

func TestJob_HasQueuedBuild_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			QueueItem: map[string]interface{}{"id": 123},
		},
		Base: "/job/test-job",
	}

	hasQueued, err := job.HasQueuedBuild(context.Background())
	assert.NoError(t, err)
	assert.True(t, hasQueued)
}

func TestJob_HasQueuedBuild_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			QueueItem: nil,
		},
		Base: "/job/test-job",
	}

	hasQueued, err := job.HasQueuedBuild(context.Background())
	assert.NoError(t, err)
	assert.False(t, hasQueued)
}

func TestJob_parentBase(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		expected string
	}{
		{
			name:     "root job",
			base:     "/job/test-job",
			expected: "",
		},
		{
			name:     "nested job",
			base:     "/job/folder/job/test-job",
			expected: "/job/folder",
		},
		{
			name:     "deeply nested job",
			base:     "/job/folder1/job/folder2/job/test-job",
			expected: "/job/folder1/job/folder2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &Job{
				Base: tt.base,
			}
			assert.Equal(t, tt.expected, job.parentBase())
		})
	}
}

func TestJob_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	status, err := job.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestJob_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	_, err := job.Poll(context.Background())
	assert.Error(t, err)
}

func TestJob_Enable_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	success, err := job.Enable(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/job/test-job/enable", mock.lastEndpoint)
}

func TestJob_Disable_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	success, err := job.Disable(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/job/test-job/disable", mock.lastEndpoint)
}

func TestJob_Delete_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	success, err := job.Delete(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/job/test-job/doDelete", mock.lastEndpoint)
}

func TestJob_Delete_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 404,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	success, err := job.Delete(context.Background())
	assert.Error(t, err)
	assert.False(t, success)
}

func TestJob_GetAllBuildIds_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		// Check that correct tree query is used
		if buildsResp, ok := response.(*struct {
			Builds []JobBuild `json:"allBuilds"`
		}); ok {
			buildsResp.Builds = []JobBuild{
				{Number: 100, URL: "http://jenkins/job/test/100/"},
				{Number: 99, URL: "http://jenkins/job/test/99/"},
				{Number: 98, URL: "http://jenkins/job/test/98/"},
			}
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	builds, err := job.GetAllBuildIds(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 3, len(builds))
	assert.Equal(t, int64(100), builds[0].Number)
	assert.Equal(t, int64(99), builds[1].Number)
}

func TestJob_GetAllBuildIds_Empty(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		// Return empty builds
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	builds, err := job.GetAllBuildIds(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(builds))
}

func TestJob_GetAllBuildIds_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	builds, err := job.GetAllBuildIds(context.Background())
	assert.Error(t, err)
	assert.Nil(t, builds)
}

func TestJob_GetConfig_Success(t *testing.T) {
	expectedConfig := `<?xml version='1.0' encoding='UTF-8'?><project></project>`

	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetXMLFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if strPtr, ok := response.(*string); ok {
			*strPtr = expectedConfig
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	config, err := job.GetConfig(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
}

func TestJob_GetConfig_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	config, err := job.GetConfig(context.Background())
	assert.Error(t, err)
	assert.Empty(t, config)
}

func TestJob_UpdateConfig_Success(t *testing.T) {
	jenkins := newMockJenkins()
	var capturedEndpoint string
	jenkins.Requester.(*MockRequester).PostXMLFunc = func(ctx context.Context, endpoint string, xml string, response interface{}, query map[string]string) (*http.Response, error) {
		capturedEndpoint = endpoint
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	newConfig := `<?xml version='1.0' encoding='UTF-8'?><project><description>Updated</description></project>`
	err := job.UpdateConfig(context.Background(), newConfig)
	assert.NoError(t, err)

	assert.Equal(t, "/job/test-job/config.xml", capturedEndpoint)
}

func TestJob_UpdateConfig_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 500,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	err := job.UpdateConfig(context.Background(), "<config/>")
	assert.Error(t, err)
}

func TestJob_InvokeSimple_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 201,
		Header:     http.Header{"Location": []string{"http://jenkins/queue/item/42/"}},
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			InQueue: false,
		},
		Base: "/job/test-job",
	}

	queueId, err := job.InvokeSimple(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), queueId)
}

func TestJob_InvokeSimple_WithParams(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 201,
		Header:     http.Header{"Location": []string{"http://jenkins/queue/item/100/"}},
	}
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if jr, ok := response.(*JobResponse); ok {
			jr.InQueue = false
			jr.Property = []struct {
				ParameterDefinitions []ParameterDefinition `json:"parameterDefinitions"`
			}{
				{
					ParameterDefinitions: []ParameterDefinition{
						{Name: "PARAM1", Type: "StringParameterDefinition"},
					},
				},
			}
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			InQueue: false,
		},
		Base: "/job/test-job",
	}

	params := map[string]string{"PARAM1": "value1"}
	queueId, err := job.InvokeSimple(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), queueId)
}

func TestJob_InvokeSimple_AlreadyQueued(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if jr, ok := response.(*JobResponse); ok {
			jr.InQueue = true
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			InQueue: true,
		},
		Base: "/job/test-job",
	}

	queueId, err := job.InvokeSimple(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), queueId) // Returns 0 when already queued
}

func TestJob_GetBuild_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if br, ok := response.(*BuildResponse); ok {
			br.Number = 42
			br.Result = "SUCCESS"
			br.URL = "http://localhost:8080/job/test-job/42/"
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			URL: "http://localhost:8080/job/test-job/",
		},
		Base: "/job/test-job",
	}
	jenkins.Server = "http://localhost:8080/"

	build, err := job.GetBuild(context.Background(), 42)
	assert.NoError(t, err)
	assert.NotNil(t, build)
	assert.Equal(t, int64(42), build.GetBuildNumber())
}

func TestJob_GetBuild_NotFound(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		return &http.Response{StatusCode: 404}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			URL: "http://localhost:8080/job/test-job/",
		},
		Base: "/job/test-job",
	}
	jenkins.Server = "http://localhost:8080/"

	build, err := job.GetBuild(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, build)
	assert.Equal(t, "404", err.Error())
}

func TestJob_GetLastBuild_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if br, ok := response.(*BuildResponse); ok {
			br.Number = 100
			br.Result = "SUCCESS"
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			LastBuild: JobBuild{Number: 100, URL: "http://localhost:8080/job/test-job/100/"},
		},
		Base: "/job/test-job",
	}

	build, err := job.GetLastBuild(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, build)
	assert.Equal(t, int64(100), build.GetBuildNumber())
}

func TestJob_GetLastBuild_NoBuilds(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		return &http.Response{StatusCode: 404}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			LastBuild: JobBuild{Number: 0, URL: ""},
		},
		Base: "/job/test-job",
	}

	build, err := job.GetLastBuild(context.Background())
	assert.Error(t, err)
	assert.Nil(t, build)
}

func TestJob_IsRunning_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if jr, ok := response.(*JobResponse); ok {
			jr.LastBuild = JobBuild{Number: 10, URL: "http://localhost:8080/job/test-job/10/"}
		}
		if br, ok := response.(*BuildResponse); ok {
			br.Number = 10
			br.Building = true
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			LastBuild: JobBuild{Number: 10, URL: "http://localhost:8080/job/test-job/10/"},
		},
		Base: "/job/test-job",
	}

	running, err := job.IsRunning(context.Background())
	assert.NoError(t, err)
	assert.True(t, running)
}

func TestJob_IsRunning_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if jr, ok := response.(*JobResponse); ok {
			jr.LastBuild = JobBuild{Number: 10, URL: "http://localhost:8080/job/test-job/10/"}
		}
		if br, ok := response.(*BuildResponse); ok {
			br.Number = 10
			br.Building = false
			br.Result = "SUCCESS"
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw: &JobResponse{
			LastBuild: JobBuild{Number: 10, URL: "http://localhost:8080/job/test-job/10/"},
		},
		Base: "/job/test-job",
	}

	running, err := job.IsRunning(context.Background())
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestJob_IsRunning_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		return nil, assert.AnError
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{},
		Base:    "/job/test-job",
	}

	running, err := job.IsRunning(context.Background())
	assert.Error(t, err)
	assert.False(t, running)
}
