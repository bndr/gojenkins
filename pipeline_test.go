// Copyright 2017 - Tessa Nordgren
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

func TestJob_GetPipelineRuns_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if runs, ok := response.(*[]PipelineRun); ok {
			*runs = []PipelineRun{
				{
					ID:     "1",
					Name:   "#1",
					Status: "SUCCESS",
					URLs: map[string]map[string]string{
						"self": {"href": "/job/test-pipeline/1/wfapi/describe"},
					},
				},
				{
					ID:     "2",
					Name:   "#2",
					Status: "FAILED",
					URLs: map[string]map[string]string{
						"self": {"href": "/job/test-pipeline/2/wfapi/describe"},
					},
				},
			}
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "test-pipeline"},
		Base:    "/job/test-pipeline",
	}

	runs, err := job.GetPipelineRuns(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(runs))
	assert.Equal(t, "1", runs[0].ID)
	assert.Equal(t, "SUCCESS", runs[0].Status)
	assert.Equal(t, "2", runs[1].ID)
	assert.Equal(t, "FAILED", runs[1].Status)
}

func TestJob_GetPipelineRuns_Empty(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		// Return empty slice
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "test-pipeline"},
		Base:    "/job/test-pipeline",
	}

	runs, err := job.GetPipelineRuns(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(runs))
}

func TestJob_GetPipelineRuns_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "test-pipeline"},
		Base:    "/job/test-pipeline",
	}

	runs, err := job.GetPipelineRuns(context.Background())
	assert.Error(t, err)
	assert.Nil(t, runs)
}

func TestJob_GetPipelineRun_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if run, ok := response.(*PipelineRun); ok {
			run.ID = "42"
			run.Name = "#42"
			run.Status = "SUCCESS"
			run.Duration = 12345
			run.URLs = map[string]map[string]string{
				"self": {"href": "/job/test-pipeline/42/wfapi/describe"},
			}
			run.Stages = []PipelineNode{
				{
					ID:     "1",
					Name:   "Build",
					Status: "SUCCESS",
					URLs: map[string]map[string]string{
						"self": {"href": "/job/test-pipeline/42/execution/node/1/wfapi/describe"},
					},
				},
				{
					ID:     "2",
					Name:   "Test",
					Status: "SUCCESS",
					URLs: map[string]map[string]string{
						"self": {"href": "/job/test-pipeline/42/execution/node/2/wfapi/describe"},
					},
				},
			}
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "test-pipeline"},
		Base:    "/job/test-pipeline",
	}

	run, err := job.GetPipelineRun(context.Background(), "42")
	assert.NoError(t, err)
	assert.NotNil(t, run)
	assert.Equal(t, "42", run.ID)
	assert.Equal(t, "#42", run.Name)
	assert.Equal(t, "SUCCESS", run.Status)
	assert.Equal(t, int64(12345), run.Duration)
	assert.Equal(t, 2, len(run.Stages))
	assert.Equal(t, "Build", run.Stages[0].Name)
	assert.Equal(t, "Test", run.Stages[1].Name)
}

func TestJob_GetPipelineRun_NotFound(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "test-pipeline"},
		Base:    "/job/test-pipeline",
	}

	run, err := job.GetPipelineRun(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, run)
}

func TestPipelineRun_GetPendingInputActions_HasActions(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if actions, ok := response.(*[]PipelineInputAction); ok {
			*actions = []PipelineInputAction{
				{
					ID:         "input-1",
					Message:    "Deploy to production?",
					ProceedURL: "/job/pipeline/1/input/input-1/proceed",
					AbortURL:   "/job/pipeline/1/input/input-1/abort",
				},
			}
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "pipeline"},
		Base:    "/job/pipeline",
	}

	run := &PipelineRun{
		Job:  job,
		Base: "/job/pipeline/1",
		ID:   "1",
	}

	actions, err := run.GetPendingInputActions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actions))
	assert.Equal(t, "input-1", actions[0].ID)
	assert.Equal(t, "Deploy to production?", actions[0].Message)
}

func TestPipelineRun_GetPendingInputActions_NoActions(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		// Return empty slice
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "pipeline"},
		Base:    "/job/pipeline",
	}

	run := &PipelineRun{
		Job:  job,
		Base: "/job/pipeline/1",
		ID:   "1",
	}

	actions, err := run.GetPendingInputActions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(actions))
}

func TestPipelineRun_GetPendingInputActions_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "pipeline"},
		Base:    "/job/pipeline",
	}

	run := &PipelineRun{
		Job:  job,
		Base: "/job/pipeline/1",
		ID:   "1",
	}

	actions, err := run.GetPendingInputActions(context.Background())
	assert.Error(t, err)
	assert.Nil(t, actions)
}

func TestPipelineRun_GetArtifacts_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "pipeline"},
		Base:    "/job/pipeline",
	}

	run := &PipelineRun{
		Job:  job,
		Base: "/job/pipeline/1",
		ID:   "1",
	}

	artifacts, err := run.GetArtifacts(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)
}

func TestPipelineRun_GetNode_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if node, ok := response.(*PipelineNode); ok {
			node.ID = "5"
			node.Name = "Deploy"
			node.Status = "IN_PROGRESS"
			node.Duration = 5000
		}
		return &http.Response{StatusCode: 200}, nil
	}

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "pipeline"},
		Base:    "/job/pipeline",
	}

	run := &PipelineRun{
		Job:  job,
		Base: "/job/pipeline/1",
		ID:   "1",
	}

	node, err := run.GetNode(context.Background(), "5")
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "5", node.ID)
	assert.Equal(t, "Deploy", node.Name)
	assert.Equal(t, "IN_PROGRESS", node.Status)
}

func TestPipelineRun_GetNode_NotFound(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	job := &Job{
		Jenkins: jenkins,
		Raw:     &JobResponse{Name: "pipeline"},
		Base:    "/job/pipeline",
	}

	run := &PipelineRun{
		Job:  job,
		Base: "/job/pipeline/1",
		ID:   "1",
	}

	node, err := run.GetNode(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestPipelineRun_Update(t *testing.T) {
	run := &PipelineRun{
		ID:     "1",
		Name:   "#1",
		Status: "SUCCESS",
		URLs: map[string]map[string]string{
			"self": {"href": "/job/test-pipeline/1/wfapi/describe"},
		},
		Stages: []PipelineNode{
			{
				ID:     "10",
				Name:   "Build",
				Status: "SUCCESS",
				URLs: map[string]map[string]string{
					"self": {"href": "/job/test-pipeline/1/execution/node/10/wfapi/describe"},
				},
			},
		},
	}

	run.update()

	assert.Equal(t, "/job/test-pipeline/1", run.Base)
	assert.Equal(t, run, run.Stages[0].Run)
	assert.Equal(t, "/job/test-pipeline/1/execution/node/10", run.Stages[0].Base)
}
