package main

import (
	"strconv"
)

type Queue struct {
	Jenkins *Jenkins
	Raw     *queueResponse
	Base    string
}

type queueResponse struct {
	Items []taskResponse
}

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
	Parameters []Parameter
}

func (q *Queue) Tasks() []*Task {
	tasks := make([]*Task, len(q.Raw.Items))
	for i, t := range q.Raw.Items {
		tasks[i] = &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t}
	}
	return tasks
}

func (q *Queue) GetTaskById(id int64) *Task {
	for _, t := range q.Raw.Items {
		if t.ID == id {
			return &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t}
		}
	}
	return nil
}

func (q *Queue) GetTasksForJob(name string) []*Task {
	tasks := make([]*Task, 0)
	for _, t := range q.Raw.Items {
		if t.Task.Name == name {
			tasks = append(tasks, &Task{Jenkins: q.Jenkins, Queue: q, Raw: &t})
		}
	}
	return tasks
}

func (q *Queue) CancelTask(id int64) bool {
	task := q.GetTaskById(id)
	return task.Cancel()
}

func (t *Task) Cancel() bool {
	qr := map[string]string{
		"id": strconv.FormatInt(t.Raw.ID, 10),
	}
	t.Jenkins.Requester.Post(t.Jenkins.GetQueueUrl()+"/cancelItem", nil, t.Raw, qr)
	return t.Jenkins.Requester.LastResponse.StatusCode == 200
}

func (t *Task) GetJob() *Job {
	return t.Jenkins.GetJob(t.Raw.Task.Name)
}

func (t *Task) GetParameters() []Parameter {
	for _, a := range t.Raw.Actions {
		if a.Parameters != nil {
			return a.Parameters
		}
	}
	return nil
}

func (t *Task) GetCauses() []map[string]interface{} {
	for _, a := range t.Raw.Actions {
		if a.Causes != nil {
			return a.Causes
		}
	}
	return nil
}

func (q *Queue) Poll() int {
	q.Jenkins.Requester.GetJSON(q.Base, q.Raw, nil)
	return q.Jenkins.Requester.LastResponse.StatusCode
}
