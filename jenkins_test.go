package gojenkins

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	J *Jenkins
)

func TestGetPlugins(t *testing.T) {
	ctx := GetTestContext()
	plugins, err := J.GetPlugins(ctx, 3)
	require.NoError(t, err)
	require.Greater(t, plugins.Count(), 0)
}

func TestGetViews(t *testing.T) {
	ctx := GetTestContext()
	views := []string{
		"View1",
		"View2",
		"View3",
	}

	defer cleanupViews(ctx, views)
	for _, view := range views {
		_, err := J.CreateView(ctx, view, LIST_VIEW)
		require.NoError(t, err)
	}

	viewsObj, _ := J.GetAllViews(ctx)

	// + 1 because of the default all view.
	assert.Equal(t, len(viewsObj), len(views)+1)
	assert.Equal(t, len(viewsObj[0].Raw.Jobs), 0)
}

func TestCreateViews(t *testing.T) {
	ctx := GetTestContext()

	job := createMockJob(t, ctx, "job")
	defer job.Delete(ctx)

	job2 := createMockJob(t, ctx, "job2")
	defer job2.Delete(ctx)

	list_view, err := J.CreateView(ctx, "test_list_view", LIST_VIEW)
	defer list_view.Delete(ctx)

	require.NoError(t, err)
	require.Equal(t, "test_list_view", list_view.GetName())
	require.Equal(t, "", list_view.GetDescription())
	require.Equal(t, 0, len(list_view.GetJobs()))

	my_view, err := J.CreateView(ctx, "test_my_view", MY_VIEW)
	defer my_view.Delete(ctx)
	require.NoError(t, err)
	require.Equal(t, "test_my_view", my_view.GetName())
	require.Equal(t, "", my_view.GetDescription())
	require.Equal(t, 2, len(my_view.GetJobs()))
}

func TestGetSingleView(t *testing.T) {
	ctx := GetTestContext()

	view, err := J.GetView(ctx, "all")
	require.NoError(t, err)
	require.NotNil(t, view)

	assert.Equal(t, view.GetName(), "all")
}

func TestDeleteView(t *testing.T) {
	ctx := GetTestContext()

	job := createMockJob(t, ctx, "job")
	defer job.Delete(ctx)

	my_view, err := J.CreateView(ctx, "test_my_view", LIST_VIEW)
	require.NoError(t, err)

	err = my_view.Delete(ctx)
	assert.NoError(t, err)

	v, err := J.GetView(ctx, my_view.GetName())
	assert.Nil(t, v)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
}
func TestDeleteFolder(t *testing.T) {
	ctx := GetTestContext()
	folders := []string{"folder1_test", "folder2_test"}
	createdFolders := []*Folder{}
	for _, folder := range folders {
		folderObj, err := J.CreateFolder(ctx, folder)
		createdFolders = append(createdFolders, folderObj)
		assert.NoError(t, err)
		assert.NotNil(t, folderObj)
		assert.Equal(t, folder, folderObj.GetName())
	}

	for _, folderObj := range createdFolders {
		err := folderObj.Delete(ctx)
		assert.NoError(t, err)

		_, err = J.GetFolder(ctx, folderObj.GetName())
		assert.ErrorIs(t, err, ErrNotFound)
	}
}

func TestCreateFolder(t *testing.T) {
	ctx := GetTestContext()
	folders := []string{"folder1_test", "folder2_test"}

	for _, folder := range folders {
		folderObj, err := J.CreateFolder(ctx, folder)
		defer folderObj.Delete(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, folderObj)
		assert.Equal(t, folder, folderObj.GetName())
	}
}

func TestCreateJobInFolder(t *testing.T) {
	ctx := GetTestContext()
	jobName := "Job_test"
	jobName2 := "Job_test2"
	job_data := getFileAsString("job.xml")

	folderObj, err := J.CreateFolder(ctx, "folder1_test")
	defer folderObj.Delete(ctx)
	require.NoError(t, err)
	require.NotNil(t, folderObj)

	folder2, err := J.CreateFolder(ctx, "folder2_test", "folder1_test")
	defer folder2.Delete(ctx)
	require.NoError(t, err)
	require.NotNil(t, folder2)

	job1, err := J.CreateJobInFolder(ctx, job_data, jobName, "folder1_test")
	defer job1.Delete(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, jobName, job1.GetName())

	job2, err := J.CreateJobInFolder(ctx, job_data, jobName2, "folder1_test", "folder2_test")
	defer job2.Delete(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, jobName2, job2.GetName())
}

func TestGetFolder(t *testing.T) {
	ctx := GetTestContext()
	folderObj, err := J.CreateFolder(ctx, "folder1_test")
	defer folderObj.Delete(ctx)
	require.NoError(t, err)
	require.NotNil(t, folderObj)

	folder2, err := J.CreateFolder(ctx, "folder2_test", "folder1_test")
	require.NoError(t, err)
	require.NotNil(t, folder2)

	folder1, err := J.GetFolder(ctx, "folder1_test")
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, "folder1_test", folder1.GetName())

	folder2, err = J.GetFolder(ctx, "folder2_test", "folder1_test")
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, "folder2_test", folder2.GetName())
}
func TestInstallPlugin(t *testing.T) {
	ctx := GetTestContext()
	err := J.InstallPlugin(ctx, "packer", "1.4")
	assert.Nil(t, err, "Could not install plugin")
}

func TestConcurrentRequests(t *testing.T) {
	ctx := GetTestContext()
	for i := 0; i <= 16; i++ {
		go func() {
			J.GetAllJobs(ctx)
			J.GetAllViews(ctx)
			J.GetAllNodes(ctx)
		}()
	}
}

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

func cleanupViews(ctx context.Context, views []string) {
	for _, view := range views {
		retrievedView, _ := J.GetView(ctx, view)
		retrievedView.Delete(ctx)
	}

}
