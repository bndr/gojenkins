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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFolder_GetName(t *testing.T) {
	folder := &Folder{
		Raw: &FolderResponse{
			Name: "my-folder",
		},
	}

	assert.Equal(t, "my-folder", folder.GetName())
}

func TestFolder_GetName_Empty(t *testing.T) {
	folder := &Folder{
		Raw: &FolderResponse{
			Name: "",
		},
	}

	assert.Equal(t, "", folder.GetName())
}

func TestFolder_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "/job/my-folder",
	}

	status, err := folder.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestFolder_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "/job/my-folder",
	}

	_, err := folder.Poll(context.Background())
	assert.Error(t, err)
}

func TestFolder_Create_Success(t *testing.T) {
	jenkins := newMockJenkins()
	requester := jenkins.Requester.(*MockRequester)
	var postEndpoint string
	var pollEndpoint string
	requester.PostFunc = func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
		postEndpoint = endpoint
		assert.IsType(t, &FolderResponse{}, response)
		assert.Equal(t, "child-folder", query["name"])
		return &http.Response{StatusCode: 200}, nil
	}
	requester.GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		pollEndpoint = endpoint
		return &http.Response{StatusCode: 200}, nil
	}

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "/job/parent/job/current-folder",
	}

	result, err := folder.Create(context.Background(), "child-folder")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "/job/parent/job/current-folder/job/child-folder", result.Base)
	assert.Equal(t, "/job/parent/job/current-folder/createItem", postEndpoint)
	assert.Equal(t, result.Base, pollEndpoint)
}

func TestFolder_Create_RootFolder(t *testing.T) {
	jenkins := newMockJenkins()
	requester := jenkins.Requester.(*MockRequester)
	var postEndpoint string
	var pollEndpoint string
	requester.PostFunc = func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
		postEndpoint = endpoint
		return &http.Response{StatusCode: 200}, nil
	}
	requester.GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		pollEndpoint = endpoint
		return &http.Response{StatusCode: 200}, nil
	}

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "",
	}

	result, err := folder.Create(context.Background(), "child-folder")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "/job/child-folder", result.Base)
	assert.Equal(t, "/createItem", postEndpoint)
	assert.Equal(t, result.Base, pollEndpoint)
}

func TestJenkins_CreateFolder_NestedParents(t *testing.T) {
	jenkins := newMockJenkins()
	requester := jenkins.Requester.(*MockRequester)
	var postEndpoint string
	var pollEndpoint string
	requester.PostFunc = func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
		postEndpoint = endpoint
		assert.Equal(t, "child-folder", query["name"])
		return &http.Response{StatusCode: 200}, nil
	}
	requester.GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		pollEndpoint = endpoint
		return &http.Response{StatusCode: 200}, nil
	}

	result, err := jenkins.CreateFolder(context.Background(), "child-folder", "parent", "current-folder")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "/job/parent/job/current-folder/job/child-folder", result.Base)
	assert.Equal(t, "/job/parent/job/current-folder/createItem", postEndpoint)
	assert.Equal(t, result.Base, pollEndpoint)
}

func TestFolder_Create_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	requester := jenkins.Requester.(*MockRequester)
	var postEndpoint string
	var pollEndpoint string
	requester.PostFunc = func(ctx context.Context, endpoint string, payload io.Reader, response interface{}, query map[string]string) (*http.Response, error) {
		postEndpoint = endpoint
		return &http.Response{StatusCode: 400}, nil
	}
	requester.GetJSONFunc = func(ctx context.Context, endpoint string, response interface{}, query map[string]string) (*http.Response, error) {
		pollEndpoint = endpoint
		return &http.Response{StatusCode: 200}, nil
	}

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "/job/parent/job/current-folder",
	}

	result, err := folder.Create(context.Background(), "child-folder")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "/job/parent/job/current-folder/createItem", postEndpoint)
	assert.Equal(t, "", pollEndpoint)
}
