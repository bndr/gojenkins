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

func TestCreateJenkins_ValidURL(t *testing.T) {
	jenkins := CreateJenkins(nil, "http://jenkins.local")

	assert.Equal(t, "http://jenkins.local", jenkins.Server)
	assert.NotNil(t, jenkins.Requester)
}

func TestCreateJenkins_URLWithTrailingSlash(t *testing.T) {
	jenkins := CreateJenkins(nil, "http://jenkins.local/")

	// Trailing slash should be removed
	assert.Equal(t, "http://jenkins.local", jenkins.Server)
}

func TestCreateJenkins_WithAuth(t *testing.T) {
	jenkins := CreateJenkins(nil, "http://jenkins.local", "admin", "password123")

	assert.Equal(t, "http://jenkins.local", jenkins.Server)
	assert.NotNil(t, jenkins.Requester)
}

func TestCreateJenkins_WithCustomClient(t *testing.T) {
	customClient := &http.Client{}
	jenkins := CreateJenkins(customClient, "http://jenkins.local")

	assert.Equal(t, "http://jenkins.local", jenkins.Server)
	assert.NotNil(t, jenkins.Requester)
}

func TestJenkins_Init_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{
			StatusCode: 200,
			Header:     http.Header{"X-Jenkins": []string{"2.375"}},
		},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if exec, ok := response.(*ExecutorResponse); ok {
				exec.NodeName = "master"
				exec.NumExecutors = 2
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"X-Jenkins": []string{"2.375"}},
			}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	result, err := jenkins.Init(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "2.375", jenkins.Version)
}

func TestJenkins_Init_ConnectionFailure(t *testing.T) {
	mock := &MockRequester{
		err: assert.AnError,
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	result, err := jenkins.Init(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestJenkins_Init_InvalidCredentials(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{
			StatusCode: 401,
			Header:     http.Header{},
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
		Raw:       nil,
	}

	result, err := jenkins.Init(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestJenkins_GetJob_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if jr, ok := response.(*JobResponse); ok {
				jr.Name = "test-job"
				jr.Description = "Test description"
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	job, err := jenkins.GetJob(context.Background(), "test-job")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-job", job.GetName())
}

func TestJenkins_GetJob_NotFound(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 404},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	job, err := jenkins.GetJob(context.Background(), "nonexistent-job")
	assert.Error(t, err)
	assert.Nil(t, job)
}

func TestJenkins_GetJob_NestedJob(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if jr, ok := response.(*JobResponse); ok {
				jr.Name = "nested-job"
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	job, err := jenkins.GetJob(context.Background(), "nested-job", "parent-folder")
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "/job/parent-folder/job/nested-job", job.Base)
}

func TestJenkins_GetBuild_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if jr, ok := response.(*JobResponse); ok {
				jr.Name = "test-job"
				jr.URL = "http://jenkins.local/job/test-job/"
			}
			if br, ok := response.(*BuildResponse); ok {
				br.Number = 42
				br.Result = "SUCCESS"
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	build, err := jenkins.GetBuild(context.Background(), "test-job", 42)
	assert.NoError(t, err)
	assert.NotNil(t, build)
}

func TestJenkins_GetBuild_JobNotFound(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 404},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	build, err := jenkins.GetBuild(context.Background(), "nonexistent-job", 1)
	assert.Error(t, err)
	assert.Nil(t, build)
}

func TestJenkins_GetAllJobs_Success(t *testing.T) {
	callCount := 0
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if endpoint == "/" {
				if exec, ok := response.(*ExecutorResponse); ok {
					exec.Jobs = []InnerJob{
						{Name: "job1", Url: "http://jenkins/job/job1/"},
						{Name: "job2", Url: "http://jenkins/job/job2/"},
					}
				}
			} else {
				if jr, ok := response.(*JobResponse); ok {
					callCount++
					jr.Name = "job" + string(rune('0'+callCount))
				}
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	jobs, err := jenkins.GetAllJobs(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobs))
}

func TestJenkins_GetAllJobs_Empty(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if exec, ok := response.(*ExecutorResponse); ok {
				exec.Jobs = []InnerJob{}
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	jobs, err := jenkins.GetAllJobs(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(jobs))
}

func TestJenkins_GetAllNodes_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if computers, ok := response.(*Computers); ok {
				computers.Computers = []*NodeResponse{
					{DisplayName: "master", NumExecutors: 2, Offline: false},
					{DisplayName: "agent-1", NumExecutors: 4, Offline: false},
				}
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	nodes, err := jenkins.GetAllNodes(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodes))
	assert.Equal(t, "master", nodes[0].Raw.DisplayName)
	assert.Equal(t, "agent-1", nodes[1].Raw.DisplayName)
}

func TestJenkins_GetAllNodes_MasterOnly(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if computers, ok := response.(*Computers); ok {
				computers.Computers = []*NodeResponse{
					{DisplayName: "master", NumExecutors: 2, Offline: false},
				}
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	nodes, err := jenkins.GetAllNodes(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "master", nodes[0].Raw.DisplayName)
}

func TestJenkins_GetNode_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if nr, ok := response.(*NodeResponse); ok {
				nr.DisplayName = "agent-1"
				nr.NumExecutors = 4
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	node, err := jenkins.GetNode(context.Background(), "agent-1")
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "agent-1", node.Raw.DisplayName)
}

func TestJenkins_GetNode_NotFound(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 404},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	node, err := jenkins.GetNode(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestJenkins_Poll_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
		Raw:       &ExecutorResponse{},
	}

	status, err := jenkins.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestJenkins_Poll_Error(t *testing.T) {
	mock := &MockRequester{
		err: assert.AnError,
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
		Raw:       &ExecutorResponse{},
	}

	_, err := jenkins.Poll(context.Background())
	assert.Error(t, err)
}

func TestJenkins_Info_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{
			StatusCode: 200,
			Header:     http.Header{"X-Jenkins": []string{"2.400"}},
		},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if exec, ok := response.(*ExecutorResponse); ok {
				exec.NodeName = "master"
				exec.NumExecutors = 4
			}
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"X-Jenkins": []string{"2.400"}},
			}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
		Raw:       &ExecutorResponse{},
	}

	info, err := jenkins.Info(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "2.400", jenkins.Version)
}

func TestJenkins_DeleteJob_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	success, err := jenkins.DeleteJob(context.Background(), "test-job")
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestJenkins_DeleteJob_NotFound(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 404},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	success, err := jenkins.DeleteJob(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.False(t, success)
}

func TestJenkins_BuildJob_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{
			StatusCode: 201,
			Header:     http.Header{"Location": []string{"http://jenkins/queue/item/123/"}},
		},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if jr, ok := response.(*JobResponse); ok {
				jr.InQueue = false
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	queueId, err := jenkins.BuildJob(context.Background(), "test-job", nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), queueId)
}

func TestJenkins_GetQueue_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if qr, ok := response.(*queueResponse); ok {
				qr.Items = []taskResponse{
					{ID: 1, Why: "Waiting for executor"},
				}
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	queue, err := jenkins.GetQueue(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, queue)
}

func TestJenkins_GetView_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if vr, ok := response.(*ViewResponse); ok {
				vr.Name = "All"
				vr.Description = "All jobs"
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	view, err := jenkins.GetView(context.Background(), "All")
	assert.NoError(t, err)
	assert.NotNil(t, view)
	assert.Equal(t, "All", view.GetName())
}

func TestJenkins_GetPlugins_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
		GetJSONFunc: func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
			if pr, ok := response.(*PluginResponse); ok {
				pr.Plugins = []Plugin{
					{ShortName: "git", LongName: "Git Plugin"},
				}
			}
			return &http.Response{StatusCode: 200}, nil
		},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	plugins, err := jenkins.GetPlugins(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, plugins)
	assert.Equal(t, 1, plugins.Count())
}

func TestJenkins_SafeRestart_Success(t *testing.T) {
	mock := &MockRequester{
		response: &http.Response{StatusCode: 200},
	}

	jenkins := &Jenkins{
		Server:    "http://jenkins.local",
		Requester: mock,
	}

	err := jenkins.SafeRestart(context.Background())
	assert.NoError(t, err)
}
