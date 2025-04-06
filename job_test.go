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

func TestGetAllBuilds(t *testing.T) {
	ctx := GetTestContext()

	jobs, cleanup := createMockJobs(t, ctx, 2)
	defer cleanup()

	for _, item := range jobs {
		t.Run(item.GetDetails().FullName, func(t *testing.T) {
			_, err := item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
			assert.NoError(t, err)
			item.Poll(ctx)
			isQueued, _ := item.IsQueued(ctx)
			assert.Equal(t, true, isQueued)
			time.Sleep(10 * time.Second)
			builds, _ := item.GetAllBuildIds(ctx)

			assert.True(t, (len(builds) > 0))
		})
	}

	job := jobs[0]
	builds, err := J.GetAllBuildIds(ctx, job.GetName())
	require.NoError(t, err)
	require.Equal(t, 1, len(builds))
	for _, b := range builds {
		build, err := J.GetBuild(ctx, job.GetName(), b.Number)
		assert.NoError(t, err)
		assert.Equal(t, "SUCCESS", build.GetResult())
	}
}

func TestBuildMethods(t *testing.T) {
	ctx := GetTestContext()
	jobName := "Job1_test"
	job := createMockJob(t, ctx, jobName)
	defer job.Delete(ctx)
	// Start the job
	_, err := job.InvokeSimple(ctx, map[string]string{"params1": "param1"})
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	_, err = job.Poll(ctx)
	require.NoError(t, err)
	build, err := job.GetLastBuild(ctx)
	require.NoError(t, err)
	require.NotNil(t, build)
	params := build.GetParameters()
	require.Equal(t, "params1", params[0].Name)
}

func TestGetSingleJob(t *testing.T) {
	ctx := GetTestContext()
	jobName := "Job1_test"
	job := createMockJob(t, ctx, jobName)
	defer job.Delete(ctx)
	retrievedJob, err := J.GetJob(ctx, jobName)
	require.NoError(t, err)
	isRunning, err := retrievedJob.IsRunning(ctx)
	require.NoError(t, err)
	config, err := retrievedJob.GetConfig(ctx)
	require.NoError(t, err)

	assert.Equal(t, false, isRunning)
	assert.Contains(t, config, "<project>")
	assert.Equal(t, jobName, retrievedJob.GetName())
}

func TestEnableDisableJob(t *testing.T) {
	ctx := GetTestContext()
	jobName := "Job1_test"
	createdJob := createMockJob(t, ctx, jobName)
	defer createdJob.Delete(ctx)
	job, err := J.GetJob(ctx, "Job1_test")
	require.NoError(t, err)
	result, err := job.Disable(ctx)
	assert.NoError(t, err)
	assert.Equal(t, true, result)
	result, err = job.Enable(ctx)
	assert.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestCopyDeleteJob(t *testing.T) {
	ctx := GetTestContext()
	job := createMockJob(t, ctx, "my_job")
	defer job.Delete(ctx)
	jobCopy, _ := job.Copy(ctx, "Job1_test_copy")
	assert.Equal(t, jobCopy.GetName(), "Job1_test_copy")
	jobDelete, _ := job.Delete(ctx)
	assert.Equal(t, true, jobDelete)
}

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

func createMockJobs(t *testing.T, ctx context.Context, count int) ([]*Job, func()) {
	jobs := []*Job{}
	for i := range count {
		jobName := fmt.Sprintf("job_%d", i)
		job := createMockJob(t, ctx, jobName)
		jobs = append(jobs, job)
	}

	cleanupFunc := func() {
		for _, job := range jobs {
			job.Delete(ctx)
		}
	}
	return jobs, cleanupFunc
}
