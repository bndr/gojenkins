package gojenkins

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

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
	j, teardown := Setup(t, ctx)
	defer teardown()
	if _, err := createTestJobs(ctx, j); err != nil {
		t.Fatalf("failed to create test jobs")
	}

	jobs, _ := j.GetAllJobs(ctx)
	defer deleteJobs(ctx, jobs)
	for _, item := range jobs {
		queueID, _ = item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
		item.Poll(ctx)
		isQueued, _ := item.IsQueued(ctx)
		require.Equal(t, true, isQueued)
		time.Sleep(10 * time.Second)
		builds, _ := item.GetAllBuildIds(ctx)

		require.True(t, (len(builds) > 0))

	}
}

// func TestGetQueueItem(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	task, err := j.GetQueueItem(ctx, queueID)
// 	require.NoError(t, err)
// 	require.NotNil(t, task.Raw)
// 	require.NotNil(t, task.Raw.ID)
// }

// func TestParseBuildHistory(t *testing.T) {
// 	r, err := os.Open("_tests/build_history.txt")
// 	require.NoError(t, err)
// 	history := parseBuildHistory(r)
// 	require.True(t, len(history) == 3)
// }

// func TestCreateViews(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()

// 	jobs, err := createTestJobs(ctx, j)
// 	require.NoError(t, err)
// 	defer deleteJobs(ctx, jobs)
// 	list_view, err := j.CreateView(ctx, "test_list_view", LIST_VIEW)
// 	defer list_view.Delete(ctx)

// 	require.NoError(t, err)
// 	require.Equal(t, "test_list_view", list_view.GetName())
// 	require.Equal(t, "", list_view.GetDescription())
// 	require.Equal(t, 0, len(list_view.GetJobs()))

// 	my_view, err := j.CreateView(ctx, "test_my_view", MY_VIEW)
// 	defer my_view.Delete(ctx)
// 	require.NoError(t, err)
// 	require.Equal(t, "test_my_view", my_view.GetName())
// 	require.Equal(t, "", my_view.GetDescription())
// 	require.Equal(t, 2, len(my_view.GetJobs()))
// }

// func TestGetAllJobs(t *testing.T) {
// 	ctx := context.Background()

// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	testJobs, err := createTestJobs(ctx, j)
// 	defer deleteJobs(ctx, testJobs)
// 	require.NoError(t, err)
// 	jobs, _ := j.GetAllJobs(ctx)
// 	require.Equal(t, 2, len(jobs))
// 	require.Equal(t, jobs[0].Raw.Color, "notbuilt")
// }

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
