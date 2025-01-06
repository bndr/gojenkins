package gojenkins

import (
	"context"
	"os"
	"testing"
)

var (
	J *Jenkins
)

// import (
// 	"context"
// 	"errors"
// 	"io/ioutil"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// var (
// 	queueID int64
// )

// func TestGetPlugins(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	plugins, err := j.GetPlugins(ctx, 3)
// 	require.NoError(t, err)
// 	require.Greater(t, 0, plugins.Count())
// }

// func TestGetViews(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()

// 	views := []string{
// 		"View1",
// 		"View2",
// 		"View3",
// 	}

// 	for _, view := range views {
// 		_, err := j.CreateView(ctx, view, LIST_VIEW)
// 		require.NoError(t, err)
// 	}

// 	viewsObj, _ := j.GetAllViews(ctx)

// 	// + 1 because of the default all view.
// 	assert.Equal(t, len(viewsObj), len(views)+1)
// 	assert.Equal(t, len(viewsObj[0].Raw.Jobs), 0)
// }

// func TestGetSingleView(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()

// 	view, err := j.GetView(ctx, "all")
// 	require.NoError(t, err)
// 	require.NotNil(t, view)

// 	assert.Equal(t, view.GetName(), "all")
// }

// func TestCreateFolder(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	folders := []string{"folder1_test", "folder2_test"}

// 	for _, folder := range folders {
// 		folderObj, err := j.CreateFolder(ctx, folder)
// 		assert.NoError(t, err)
// 		assert.NotNil(t, folderObj)
// 		assert.Equal(t, folder, folderObj.GetName())
// 	}
// }

// func TestCreateJobInFolder(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	jobName := "Job_test"
// 	jobName2 := "Job_test2"
// 	job_data := getFileAsString("job.xml")

// 	folderObj, err := j.CreateFolder(ctx, "folder1_test")
// 	require.NoError(t, err)
// 	require.NotNil(t, folderObj)

// 	folder2, err := j.CreateFolder(ctx, "folder2_test", "folder1_test")
// 	require.NoError(t, err)
// 	require.NotNil(t, folder2)

// 	job1, err := j.CreateJobInFolder(ctx, job_data, jobName, "folder1_test")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, job1)
// 	assert.Equal(t, "Some Job Description", job1.GetDescription())
// 	assert.Equal(t, jobName, job1.GetName())

// 	job2, err := j.CreateJobInFolder(ctx, job_data, jobName2, "folder1_test", "folder2_test")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, job2)
// 	assert.Equal(t, "Some Job Description", job2.GetDescription())
// 	assert.Equal(t, jobName2, job2.GetName())
// }

// func TestGetFolder(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	folderObj, err := j.CreateFolder(ctx, "folder1_test")
// 	require.NoError(t, err)
// 	require.NotNil(t, folderObj)

// 	folder2, err := j.CreateFolder(ctx, "folder2_test", "folder1_test")
// 	require.NoError(t, err)
// 	require.NotNil(t, folder2)

// 	folder1, err := j.GetFolder(ctx, "folder1_test")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, folder1)
// 	assert.Equal(t, "folder1_test", folder1.GetName())

// 	folder2, err = j.GetFolder(ctx, "folder2_test", "folder1_test")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, folder2)
// 	assert.Equal(t, "folder2_test", folder2.GetName())
// }
// func TestInstallPlugin(t *testing.T) {
// 	ctx := context.Background()

// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	err := j.InstallPlugin(ctx, "packer", "1.4")

// 	assert.Nil(t, err, "Could not install plugin")
// }

// func TestConcurrentRequests(t *testing.T) {
// 	ctx := context.Background()
// 	j, teardown := Setup(t, ctx)
// 	defer teardown()
// 	for i := 0; i <= 16; i++ {
// 		go func() {
// 			j.GetAllJobs(ctx)
// 			j.GetAllViews(ctx)
// 			j.GetAllNodes(ctx)
// 		}()
// 	}
// }

func TestMain(m *testing.M) {
	ctx := context.Background()
	testContext, err := Setup(ctx, jenkinsUsername, jenkinsPassword)
	if err != nil {
		panic(err)
	}

	defer testContext.CleanupFunc()
	J = testContext.Jenkins

	// Run tests
	code := m.Run()

	// Exit with the code from the test run
	os.Exit(code)
}
