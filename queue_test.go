package gojenkins

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueTasksPreservesDistinctItems(t *testing.T) {
	q := &Queue{
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

	tasks := q.Tasks()
	require.Len(t, tasks, 3)
	assert.Equal(t, int64(1), tasks[0].Raw.ID)
	assert.Equal(t, int64(2), tasks[1].Raw.ID)
	assert.Equal(t, int64(3), tasks[2].Raw.ID)

	jobTasks := q.GetTasksForJob("job-a")
	require.Len(t, jobTasks, 2)
	assert.Equal(t, int64(1), jobTasks[0].Raw.ID)
	assert.Equal(t, int64(3), jobTasks[1].Raw.ID)
}

func TestCancelTaskMissingID(t *testing.T) {
	q := &Queue{Raw: &queueResponse{Items: []taskResponse{{ID: 1}}}}
	ok, err := q.CancelTask(context.Background(), 99)
	assert.False(t, ok)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
