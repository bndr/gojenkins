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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuild_Info(t *testing.T) {
	rawResponse := &BuildResponse{
		Number: 123,
		Result: "SUCCESS",
	}
	build := &Build{
		Raw: rawResponse,
	}

	assert.Equal(t, rawResponse, build.Info())
}

func TestBuild_GetBuildNumber(t *testing.T) {
	build := &Build{
		Raw: &BuildResponse{
			Number: 42,
		},
	}

	assert.Equal(t, int64(42), build.GetBuildNumber())
}

func TestBuild_GetResult(t *testing.T) {
	tests := []struct {
		name     string
		result   string
		expected string
	}{
		{"success", "SUCCESS", "SUCCESS"},
		{"failure", "FAILURE", "FAILURE"},
		{"unstable", "UNSTABLE", "UNSTABLE"},
		{"aborted", "ABORTED", "ABORTED"},
		{"not built", "NOT_BUILT", "NOT_BUILT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{
				Raw: &BuildResponse{
					Result: tt.result,
				},
			}
			assert.Equal(t, tt.expected, build.GetResult())
		})
	}
}

func TestBuild_GetUrl(t *testing.T) {
	build := &Build{
		Raw: &BuildResponse{
			URL: "http://jenkins/job/test-job/42/",
		},
	}

	assert.Equal(t, "http://jenkins/job/test-job/42/", build.GetUrl())
}

func TestBuild_GetCulprits(t *testing.T) {
	culprits := []Culprit{
		{FullName: "John Doe", AbsoluteUrl: "http://jenkins/user/johndoe"},
		{FullName: "Jane Doe", AbsoluteUrl: "http://jenkins/user/janedoe"},
	}
	build := &Build{
		Raw: &BuildResponse{
			Culprits: culprits,
		},
	}

	result := build.GetCulprits()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "John Doe", result[0].FullName)
	assert.Equal(t, "Jane Doe", result[1].FullName)
}

func TestBuild_GetActions(t *testing.T) {
	actions := []generalObj{
		{TotalCount: 10},
		{TotalCount: 5},
	}
	build := &Build{
		Raw: &BuildResponse{
			Actions: actions,
		},
	}

	result := build.GetActions()
	assert.Equal(t, 2, len(result))
}

func TestBuild_GetParameters(t *testing.T) {
	build := &Build{
		Raw: &BuildResponse{
			Actions: []generalObj{
				{
					Parameters: []parameter{
						{Name: "PARAM1", Value: "value1"},
						{Name: "PARAM2", Value: "value2"},
					},
				},
			},
		},
	}

	params := build.GetParameters()
	assert.Equal(t, 2, len(params))
	assert.Equal(t, "PARAM1", params[0].Name)
	assert.Equal(t, "value1", params[0].Value)
}

func TestBuild_GetParameters_NoParameters(t *testing.T) {
	build := &Build{
		Raw: &BuildResponse{
			Actions: []generalObj{},
		},
	}

	params := build.GetParameters()
	assert.Nil(t, params)
}

func TestBuild_GetTimestamp(t *testing.T) {
	// Test timestamp: 2023-01-15 10:30:00 UTC = 1673778600000 milliseconds
	build := &Build{
		Raw: &BuildResponse{
			Timestamp: 1673778600000,
		},
	}

	result := build.GetTimestamp()
	expected := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	assert.Equal(t, expected.Unix(), result.Unix())
}

func TestBuild_GetDuration(t *testing.T) {
	build := &Build{
		Raw: &BuildResponse{
			Duration: 12345.67,
		},
	}

	assert.Equal(t, 12345.67, build.GetDuration())
}

func TestBuild_IsGood_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Result:   STATUS_SUCCESS,
			Building: false,
		},
		Base: "/job/test-job/1",
	}

	assert.True(t, build.IsGood(context.Background()))
}

func TestBuild_IsGood_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Result:   STATUS_FAIL,
			Building: false,
		},
		Base: "/job/test-job/1",
	}

	assert.False(t, build.IsGood(context.Background()))
}

func TestBuild_IsGood_Running(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Result:   "",
			Building: true,
		},
		Base: "/job/test-job/1",
	}

	// A running build is not "good" yet
	assert.False(t, build.IsGood(context.Background()))
}

func TestBuild_IsRunning_True(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Building: true,
		},
		Base: "/job/test-job/1",
	}

	assert.True(t, build.IsRunning(context.Background()))
}

func TestBuild_IsRunning_False(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Building: false,
		},
		Base: "/job/test-job/1",
	}

	assert.False(t, build.IsRunning(context.Background()))
}

func TestBuild_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw:     &BuildResponse{},
		Base:    "/job/test-job/1",
		Depth:   1,
	}

	status, err := build.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestBuild_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	build := &Build{
		Jenkins: jenkins,
		Raw:     &BuildResponse{},
		Base:    "/job/test-job/1",
		Depth:   1,
	}

	_, err := build.Poll(context.Background())
	assert.Error(t, err)
}

func TestBuild_GetArtifacts(t *testing.T) {
	jenkins := newMockJenkins()
	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Artifacts: []struct {
				DisplayPath  string `json:"displayPath"`
				FileName     string `json:"fileName"`
				RelativePath string `json:"relativePath"`
			}{
				{FileName: "artifact1.jar", RelativePath: "target/artifact1.jar"},
				{FileName: "artifact2.jar", RelativePath: "target/artifact2.jar"},
			},
		},
		Base: "/job/test-job/1",
	}

	artifacts := build.GetArtifacts()
	assert.Equal(t, 2, len(artifacts))
	assert.Equal(t, "artifact1.jar", artifacts[0].FileName)
	assert.Equal(t, "artifact2.jar", artifacts[1].FileName)
	assert.Equal(t, "/job/test-job/1/artifact/target/artifact1.jar", artifacts[0].Path)
}

func TestBuild_GetRevision_Git(t *testing.T) {
	build := &Build{
		Raw: &BuildResponse{
			ChangeSet: struct {
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
			}{
				Kind: "git",
			},
			Actions: []generalObj{
				{
					LastBuiltRevision: BuildRevision{
						SHA1: "abc123def456",
					},
				},
			},
		},
	}

	assert.Equal(t, "abc123def456", build.GetRevision())
}

func TestBuild_Stop_AlreadyStopped(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Building: false,
		},
		Base: "/job/test-job/1",
	}

	stopped, err := build.Stop(context.Background())
	assert.NoError(t, err)
	assert.True(t, stopped)
}

func TestBuild_Stop_Running(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Building: true,
		},
		Base: "/job/test-job/1",
	}

	stopped, err := build.Stop(context.Background())
	assert.NoError(t, err)
	assert.True(t, stopped)

	// Verify the correct endpoint was called
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/job/test-job/1/stop", mock.lastEndpoint)
}

func TestBuild_GetConsoleOutput_Success(t *testing.T) {
	expectedOutput := "Building...\nCompiling sources...\nBuild successful!"

	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetXMLFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if strPtr, ok := response.(*string); ok {
			*strPtr = expectedOutput
		}
		return &http.Response{StatusCode: 200}, nil
	}

	build := &Build{
		Jenkins: jenkins,
		Raw:     &BuildResponse{},
		Base:    "/job/test-job/42",
	}

	output := build.GetConsoleOutput(context.Background())
	assert.Equal(t, expectedOutput, output)
}

func TestBuild_GetConsoleOutputFromIndex_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if strPtr, ok := response.(*string); ok {
			*strPtr = "More log content..."
		}
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"X-Text-Size": []string{"1024"},
				"X-More-Data": []string{"true"},
			},
		}, nil
	}

	build := &Build{
		Jenkins: jenkins,
		Raw:     &BuildResponse{},
		Base:    "/job/test-job/42",
	}

	resp, err := build.GetConsoleOutputFromIndex(context.Background(), 512)
	assert.NoError(t, err)
	assert.Equal(t, int64(1024), resp.Offset)
	assert.True(t, resp.HasMoreText)
	assert.Equal(t, "More log content...", resp.Content)
}

func TestBuild_GetConsoleOutputFromIndex_NoMoreData(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if strPtr, ok := response.(*string); ok {
			*strPtr = "Final log content"
		}
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"X-Text-Size": []string{"2048"},
			},
		}, nil
	}

	build := &Build{
		Jenkins: jenkins,
		Raw:     &BuildResponse{},
		Base:    "/job/test-job/42",
	}

	resp, err := build.GetConsoleOutputFromIndex(context.Background(), 1024)
	assert.NoError(t, err)
	assert.Equal(t, int64(2048), resp.Offset)
	assert.False(t, resp.HasMoreText)
}

func TestBuild_GetCauses_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Actions: []generalObj{
				{
					Causes: []map[string]interface{}{
						{"shortDescription": "Started by user admin"},
					},
				},
			},
		},
		Base: "/job/test-job/42",
	}

	causes, err := build.GetCauses(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(causes))
	assert.Equal(t, "Started by user admin", causes[0]["shortDescription"])
}

func TestBuild_GetCauses_NoCauses(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	build := &Build{
		Jenkins: jenkins,
		Raw: &BuildResponse{
			Actions: []generalObj{},
		},
		Base: "/job/test-job/42",
	}

	causes, err := build.GetCauses(context.Background())
	assert.Error(t, err)
	assert.Nil(t, causes)
	assert.Contains(t, err.Error(), "no causes")
}
