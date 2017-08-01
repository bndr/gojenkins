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
	"strconv"
)

// Queue represents a queue
type Queue struct {
	Jenkins *Jenkins
	Raw     *queueResponse
	Base    string
}

type queueResponse struct {
	Items []taskResponse
}

// Task represents a task
type Task struct {
	Raw     *taskResponse
	Jenkins *Jenkins
	Queue   *Queue
}

type taskResponse struct {
	Actions                    []generalAction `json:"actions"`
	Blocked                    bool            `json:"blocked"`
	Buildable                  bool            `json:"buildable"`
	BuildableStartMilliseconds int64           `json:"buildableStartMilliseconds"`
	ID                         int64           `json:"id"`
	InQueueSince               int64           `json:"inQueueSince"`
	Params                     string          `json:"params"`
	Pending                    bool            `json:"pending"`
	Stuck                      bool            `json:"stuck"`
	Task                       struct {
		Color string `json:"color"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	} `json:"task"`
	URL string `json:"url"`
	Why string `json:"why"`
}

type generalAction struct {
	Causes     []map[string]interface{}
	Parameters []parameter
}

// Tasks gets tasks on a queue
func (q *Queue) Tasks() []*Task {
	tasks := make([]*Task, len(q.Raw.Items))
	for i, t := range q.Raw.Items {
		tasks[i] = &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t}
	}
	return tasks
}

// GetTaskById gets a task by id
func (q *Queue) GetTaskById(id int64) *Task {
	for _, t := range q.Raw.Items {
		if t.ID == id {
			return &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t}
		}
	}
	return nil
}

// GetTasksForJob gets tasks for a job
func (q *Queue) GetTasksForJob(name string) []*Task {
	tasks := make([]*Task, 0)
	for _, t := range q.Raw.Items {
		if t.Task.Name == name {
			tasks = append(tasks, &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t})
		}
	}
	return tasks
}

// CancelTask cancels a task by id
func (q *Queue) CancelTask(id int64) (bool, error) {
	task := q.GetTaskById(id)
	return task.Cancel()
}

// Cancel cancels the current task
func (t *Task) Cancel() (bool, error) {
	qr := map[string]string{
		"id": strconv.FormatInt(t.Raw.ID, 10),
	}
	response, err := t.Jenkins.Requester.Post(t.Jenkins.GetQueueUrl()+"/cancelItem", nil, t.Raw, qr)
	if err != nil {
		return false, err
	}
	return response.StatusCode == 200, nil
}

// GetJob gets the job associated with a task
func (t *Task) GetJob() (*Job, error) {
	return t.Jenkins.GetJob(t.Raw.Task.Name)
}

// GetWhy gets the why a task was queued
func (t *Task) GetWhy() string {
	return t.Raw.Why
}

// GetParameters gets a task's parameters
func (t *Task) GetParameters() []parameter {
	for _, a := range t.Raw.Actions {
		if a.Parameters != nil {
			return a.Parameters
		}
	}
	return nil
}

// GetCauses returns a task's causes
func (t *Task) GetCauses() []map[string]interface{} {
	for _, a := range t.Raw.Actions {
		if a.Causes != nil {
			return a.Causes
		}
	}
	return nil
}

// Poll polls the queue
func (q *Queue) Poll() (int, error) {
	response, err := q.Jenkins.Requester.GetJSON(q.Base, q.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
