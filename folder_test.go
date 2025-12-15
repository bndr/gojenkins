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

func TestFolder_parentBase(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		expected string
	}{
		{
			name:     "root folder",
			base:     "/job/my-folder",
			expected: "",
		},
		{
			name:     "nested folder",
			base:     "/job/parent/job/child",
			expected: "/job/parent",
		},
		{
			name:     "deeply nested folder",
			base:     "/job/level1/job/level2/job/level3",
			expected: "/job/level1/job/level2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder := &Folder{
				Base: tt.base,
			}
			assert.Equal(t, tt.expected, folder.parentBase())
		})
	}
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
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "/job/parent/job/new-folder",
	}

	result, err := folder.Create(context.Background(), "new-folder")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFolder_Create_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 400,
	}

	folder := &Folder{
		Jenkins: jenkins,
		Raw:     &FolderResponse{},
		Base:    "/job/parent/job/new-folder",
	}

	result, err := folder.Create(context.Background(), "new-folder")
	assert.Error(t, err)
	assert.Nil(t, result)
}
