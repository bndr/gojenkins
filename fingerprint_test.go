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

func TestFingerPrint_Valid_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if fp, ok := response.(*FingerPrintResponse); ok {
			fp.Hash = "abc123def456"
			fp.FileName = "artifact.jar"
		}
		return &http.Response{StatusCode: 200}, nil
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123def456",
		Raw:     &FingerPrintResponse{},
	}

	valid, err := fingerprint.Valid(context.Background())
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestFingerPrint_Valid_HashMismatch(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if fp, ok := response.(*FingerPrintResponse); ok {
			fp.Hash = "different-hash"
		}
		return &http.Response{StatusCode: 200}, nil
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123def456",
		Raw:     &FingerPrintResponse{},
	}

	valid, err := fingerprint.Valid(context.Background())
	assert.Error(t, err)
	assert.False(t, valid)
	assert.Contains(t, err.Error(), "Invalid")
}

func TestFingerPrint_Valid_NotFound(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		return &http.Response{StatusCode: 404}, nil
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "nonexistent",
		Raw:     &FingerPrintResponse{},
	}

	valid, err := fingerprint.Valid(context.Background())
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestFingerPrint_Valid_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123",
		Raw:     &FingerPrintResponse{},
	}

	valid, err := fingerprint.Valid(context.Background())
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestFingerPrint_ValidateForBuild_ValidFingerprint(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if fp, ok := response.(*FingerPrintResponse); ok {
			fp.Hash = "abc123"
			fp.FileName = "artifact.jar"
			fp.Original.Name = "test-job"
			fp.Original.Number = 42
		}
		return &http.Response{StatusCode: 200}, nil
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123",
		Raw:     &FingerPrintResponse{},
	}

	job := &Job{
		Raw: &JobResponse{Name: "test-job"},
	}
	build := &Build{
		Job: job,
		Raw: &BuildResponse{Number: 42},
	}

	valid, err := fingerprint.ValidateForBuild(context.Background(), "artifact.jar", build)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestFingerPrint_ValidateForBuild_FilenameMismatch(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if fp, ok := response.(*FingerPrintResponse); ok {
			fp.Hash = "different" // Will cause Valid() to fail
			fp.FileName = "other-artifact.jar"
		}
		return &http.Response{StatusCode: 200}, nil
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123",
		Raw:     &FingerPrintResponse{},
	}

	valid, err := fingerprint.ValidateForBuild(context.Background(), "artifact.jar", nil)
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestFingerPrint_GetInfo_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		if fp, ok := response.(*FingerPrintResponse); ok {
			fp.Hash = "abc123def456"
			fp.FileName = "artifact.jar"
			fp.Timestamp = 1673778600000
		}
		return &http.Response{StatusCode: 200}, nil
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123def456",
		Raw:     &FingerPrintResponse{},
	}

	info, err := fingerprint.GetInfo(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "abc123def456", info.Hash)
	assert.Equal(t, "artifact.jar", info.FileName)
	assert.Equal(t, int64(1673778600000), info.Timestamp)
}

func TestFingerPrint_GetInfo_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123",
		Raw:     &FingerPrintResponse{},
	}

	info, err := fingerprint.GetInfo(context.Background())
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestFingerPrint_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123",
		Raw:     &FingerPrintResponse{},
	}

	status, err := fingerprint.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	// Verify correct endpoint
	mock := jenkins.Requester.(*MockRequester)
	assert.Equal(t, "/fingerprint/abc123", mock.lastEndpoint)
}

func TestFingerPrint_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	fingerprint := FingerPrint{
		Jenkins: jenkins,
		Base:    "/fingerprint/",
		Id:      "abc123",
		Raw:     &FingerPrintResponse{},
	}

	_, err := fingerprint.Poll(context.Background())
	assert.Error(t, err)
}
