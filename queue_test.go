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

func TestQueue_Tasks(t *testing.T) {
	jenkins := newMockJenkins()
	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 1, Stuck: false, Blocked: false},
				{ID: 2, Stuck: true, Blocked: false},
				{ID: 3, Stuck: false, Blocked: true},
			},
		},
	}

	tasks := queue.Tasks()
	assert.Equal(t, 3, len(tasks))
	assert.Equal(t, int64(1), tasks[0].Raw.ID)
	assert.Equal(t, int64(2), tasks[1].Raw.ID)
	assert.Equal(t, int64(3), tasks[2].Raw.ID)
}

func TestQueue_Tasks_Empty(t *testing.T) {
	jenkins := newMockJenkins()
	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{},
		},
	}

	tasks := queue.Tasks()
	assert.Equal(t, 0, len(tasks))
}

func TestQueue_GetTaskById_Found(t *testing.T) {
	jenkins := newMockJenkins()
	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
		},
	}

	task := queue.GetTaskById(2)
	assert.NotNil(t, task)
	assert.Equal(t, int64(2), task.Raw.ID)
}

func TestQueue_GetTaskById_NotFound(t *testing.T) {
	jenkins := newMockJenkins()
	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 1},
				{ID: 2},
			},
		},
	}

	task := queue.GetTaskById(99)
	assert.Nil(t, task)
}

func TestQueue_GetTasksForJob(t *testing.T) {
	jenkins := newMockJenkins()
	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 1, Task: struct {
					Color string `json:"color"`
					Name  string `json:"name"`
					URL   string `json:"url"`
				}{Name: "job-a"}},
				{ID: 2, Task: struct {
					Color string `json:"color"`
					Name  string `json:"name"`
					URL   string `json:"url"`
				}{Name: "job-b"}},
				{ID: 3, Task: struct {
					Color string `json:"color"`
					Name  string `json:"name"`
					URL   string `json:"url"`
				}{Name: "job-a"}},
			},
		},
	}

	tasks := queue.GetTasksForJob("job-a")
	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, int64(1), tasks[0].Raw.ID)
	assert.Equal(t, int64(3), tasks[1].Raw.ID)
}

func TestQueue_GetTasksForJob_NotFound(t *testing.T) {
	jenkins := newMockJenkins()
	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 1, Task: struct {
					Color string `json:"color"`
					Name  string `json:"name"`
					URL   string `json:"url"`
				}{Name: "job-a"}},
			},
		},
	}

	tasks := queue.GetTasksForJob("nonexistent-job")
	assert.Equal(t, 0, len(tasks))
}

func TestQueue_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	queue := &Queue{
		Jenkins: jenkins,
		Raw:     &queueResponse{},
		Base:    "/queue",
	}

	status, err := queue.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestQueue_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	queue := &Queue{
		Jenkins: jenkins,
		Raw:     &queueResponse{},
		Base:    "/queue",
	}

	_, err := queue.Poll(context.Background())
	assert.Error(t, err)
}

func TestTask_GetWhy(t *testing.T) {
	task := &Task{
		Raw: &taskResponse{
			Why: "Waiting for next available executor",
		},
	}

	assert.Equal(t, "Waiting for next available executor", task.GetWhy())
}

func TestTask_GetParameters(t *testing.T) {
	task := &Task{
		Raw: &taskResponse{
			Actions: []generalAction{
				{
					Parameters: []parameter{
						{Name: "BRANCH", Value: "main"},
						{Name: "ENV", Value: "production"},
					},
				},
			},
		},
	}

	params := task.GetParameters()
	assert.Equal(t, 2, len(params))
	assert.Equal(t, "BRANCH", params[0].Name)
	assert.Equal(t, "main", params[0].Value)
}

func TestTask_GetParameters_NoParameters(t *testing.T) {
	task := &Task{
		Raw: &taskResponse{
			Actions: []generalAction{},
		},
	}

	params := task.GetParameters()
	assert.Nil(t, params)
}

func TestTask_GetCauses(t *testing.T) {
	task := &Task{
		Raw: &taskResponse{
			Actions: []generalAction{
				{
					Causes: []map[string]interface{}{
						{"shortDescription": "Started by user admin"},
					},
				},
			},
		},
	}

	causes := task.GetCauses()
	assert.Equal(t, 1, len(causes))
	assert.Equal(t, "Started by user admin", causes[0]["shortDescription"])
}

func TestTask_GetCauses_NoCauses(t *testing.T) {
	task := &Task{
		Raw: &taskResponse{
			Actions: []generalAction{},
		},
	}

	causes := task.GetCauses()
	assert.Nil(t, causes)
}

func TestTask_Cancel_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	task := &Task{
		Jenkins: jenkins,
		Raw: &taskResponse{
			ID: 123,
		},
	}

	success, err := task.Cancel(context.Background())
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestTask_Cancel_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 404,
	}

	task := &Task{
		Jenkins: jenkins,
		Raw: &taskResponse{
			ID: 123,
		},
	}

	success, err := task.Cancel(context.Background())
	assert.NoError(t, err)
	assert.False(t, success)
}

func TestTask_Poll_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	task := &Task{
		Jenkins: jenkins,
		Raw:     &taskResponse{},
		Base:    "/queue/item/123",
	}

	status, err := task.Poll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

func TestTask_Poll_Error(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).err = assert.AnError

	task := &Task{
		Jenkins: jenkins,
		Raw:     &taskResponse{},
		Base:    "/queue/item/123",
	}

	_, err := task.Poll(context.Background())
	assert.Error(t, err)
}

func TestQueue_CancelTask_Success(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 200,
	}

	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 42, Stuck: false, Blocked: false},
			},
		},
	}

	success, err := queue.CancelTask(context.Background(), 42)
	assert.NoError(t, err)
	assert.True(t, success)
}

func TestQueue_CancelTask_Failure(t *testing.T) {
	jenkins := newMockJenkins()
	jenkins.Requester.(*MockRequester).response = &http.Response{
		StatusCode: 403,
	}

	queue := &Queue{
		Jenkins: jenkins,
		Raw: &queueResponse{
			Items: []taskResponse{
				{ID: 42, Stuck: false, Blocked: false},
			},
		},
	}

	success, err := queue.CancelTask(context.Background(), 42)
	assert.NoError(t, err)
	assert.False(t, success)
}
