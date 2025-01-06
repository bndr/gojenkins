package gojenkins

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJobs(t *testing.T) {
	job1ID := "Job1_test"
	job2ID := "job2_test"
	ctx := GetTestContext()

	job1 := createMockJob(t, ctx, job1ID)
	defer job1.Delete(ctx)
	job2 := createMockJob(t, ctx, job2ID)
	defer job2.Delete(ctx)

	require.NotNil(t, job1)
	require.Equal(t, "Some Job Description", job1.GetDescription())
	require.Equal(t, job1ID, job1.GetName())

	defer job2.Delete(ctx)
	require.NotNil(t, job2)
	require.Equal(t, "Some Job Description", job2.GetDescription())
	require.Equal(t, job2ID, job2.GetName())
}

func TestCreateBuilds(t *testing.T) {
	ctx := context.Background()
	jobstr := []string{
		"job1", "job2",
	}

	jobs := []*Job{}

	for _, job := range jobstr {
		jo := createMockJob(t, ctx, job)
		jobs = append(jobs, jo)
	}

	for _, item := range jobs {
		t.Run(item.GetName(), func(t *testing.T) {
			defer item.Delete(ctx)
			queueID, err := item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
			assert.NoError(t, err)
			require.NotEqual(t, 0, queueID)

			item.Poll(ctx)
			isQueued, _ := item.IsQueued(ctx)
			require.Equal(t, true, isQueued)
			time.Sleep(10 * time.Second)
			builds, _ := item.GetAllBuildIds(ctx)

			require.True(t, (len(builds) > 0))
		})

	}
}

func TestGetQueueItem(t *testing.T) {
	ctx := GetTestContext()
	job := createMockJob(t, ctx, "mock_job")
	defer job.Delete(ctx)

	queueID, err := job.InvokeSimple(ctx, map[string]string{"params1": "param1"})
	require.NoError(t, err)
	task, err := J.GetQueueItem(ctx, queueID)
	require.NoError(t, err)
	require.NotNil(t, task.Raw)
	require.NotNil(t, task.Raw.ID)
}

func TestParseBuildHistory(t *testing.T) {
	r, err := os.Open("_tests/build_history.txt")
	require.NoError(t, err)
	history := parseBuildHistory(r)
	require.True(t, len(history) == 3)
}

func TestGetAllJobs(t *testing.T) {
	ctx := context.Background()
	createdJobs := []*Job{}
	jobsToCreate := 10
	for i := range jobsToCreate {
		newJobName := fmt.Sprintf("job_%d", i)
		job := createMockJob(t, ctx, newJobName)
		createdJobs = append(createdJobs, job)
	}

	defer func() {
		for _, job := range createdJobs {
			job.Delete(ctx)
		}
	}()

	jobs, _ := J.GetAllJobs(ctx)
	require.Equal(t, jobsToCreate, len(jobs))
	require.Equal(t, jobs[0].Raw.Color, "notbuilt")
}

// func TestGetAllBuilds(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	createdJobs, err := createTestJobs(ctx, j)
// 	defer deleteJobs(ctx, createdJobs)

// 	require.NoError(t, err)

// 	jobs, _ := j.GetAllJobs(ctx)
// 	for _, item := range jobs {
// 		queueID, _ = item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
// 		item.Poll(ctx)
// 		isQueued, _ := item.IsQueued(ctx)
// 		assert.Equal(t, true, isQueued)
// 		time.Sleep(10 * time.Second)
// 		builds, _ := item.GetAllBuildIds(ctx)

// 		assert.True(t, (len(builds) > 0))

// 	}
// 	builds, _ := j.GetAllBuildIds(ctx, "Job1_test")
// 	require.Equal(t, 1, len(builds))
// 	for _, b := range builds {
// 		build, _ := j.GetBuild(ctx, "Job1_test", b.Number)
// 		assert.Equal(t, "SUCCESS", build.GetResult())
// 	}
// }

// func TestBuildMethods(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	jobs, err := createTestJobs(ctx, j)
// 	defer deleteJobs(ctx, jobs)
// 	require.NoError(t, err)
// 	job, _ := j.GetJob(ctx, "Job1_test")
// 	require.NotNil(t, job)

// 	// Start the job
// 	_, err = job.InvokeSimple(ctx, map[string]string{})
// 	require.NoError(t, err)
// 	build, _ := job.GetLastBuild(ctx)
// 	require.NotNil(t, build)
// 	params := build.GetParameters()
// 	require.Equal(t, "params1", params[0].Name)
// }

// func TestGetSingleJob(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	jobs, err := createTestJobs(ctx, j)
// 	require.NoError(t, err)
// 	defer deleteJobs(ctx, jobs)
// 	job, _ := j.GetJob(ctx, "Job1_test")
// 	isRunning, _ := job.IsRunning(ctx)
// 	config, err := job.GetConfig(ctx)
// 	require.NoError(t, err)
// 	require.Equal(t, false, isRunning)
// 	require.Contains(t, config, "<project>")
// }

// func TestEnableDisableJob(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	_, err := createTestJobs(ctx, j)
// 	require.NoError(t, err)
// 	job, _ := j.GetJob(ctx, "Job1_test")
// 	result, _ := job.Disable(ctx)
// 	assert.Equal(t, true, result)
// 	result, _ = job.Enable(ctx)
// 	assert.Equal(t, true, result)
// }

// func TestCopyDeleteJob(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	_, err := createTestJobs(ctx, j)
// 	require.NoError(t, err)
// 	job, _ := j.GetJob(ctx, "Job1_test")
// 	jobCopy, _ := job.Copy(ctx, "Job1_test_copy")
// 	assert.Equal(t, jobCopy.GetName(), "Job1_test_copy")
// 	jobDelete, _ := job.Delete(ctx)
// 	assert.Equal(t, true, jobDelete)
// }

func getFileAsString(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

func createMockJob(t *testing.T, ctx context.Context, name string) *Job {
	t.Helper()
	jobConfig := getFileAsString("job.xml")
	job, err := J.CreateJob(ctx, jobConfig, name)
	require.NoError(t, err)
	return job
}
