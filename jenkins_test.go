package gojenkins

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	jenkins *Jenkins
	queueID int64
)

func TestInit(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	jenkins = CreateJenkins(nil, "http://localhost:8080", "admin", "admin")
	_, err := jenkins.Init(ctx)
	assert.Nil(t, err, "Jenkins Initialization should not fail")
}

func TestCreateJobs(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	job1ID := "Job1_test"
	job2ID := "job2_test"
	job_data := getFileAsString("job.xml")

	ctx := context.Background()
	job1, err := jenkins.CreateJob(ctx, job_data, job1ID)
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, job1ID, job1.GetName())

	job2, _ := jenkins.CreateJob(ctx, job_data, job2ID)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, job2ID, job2.GetName())
}

func TestCreateNodes(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	id1 := "node1_test"
	//id2 := "node2_test"
	id3 := "node3_test"
	id4 := "node4_test"

	jnlp := map[string]string{"method": "JNLPLauncher"}
	//ssh := map[string]string{"method": "SSHLauncher"}

	ctx := context.Background()
	node1, _ := jenkins.CreateNode(ctx, id1, 1, "Node 1 Description", "/var/lib/jenkins", "", jnlp)
	assert.Equal(t, id1, node1.GetName())

	//node2, _ := jenkins.CreateNode(id2, 1, "Node 2 Description", "/var/lib/jenkins", "jdk8 docker", ssh)
	//assert.Equal(t, id2, node2.GetName())

	node3, _ := jenkins.CreateNode(ctx, id3, 1, "Node 3 Description", "/var/lib/jenkins", "jdk7")
	assert.Equal(t, id3, node3.GetName())
	node4, _ := jenkins.CreateNode(ctx, id4, 1, "Node 4 Description", "/var/lib/jenkins", "jdk7")
	assert.Equal(t, id4, node4.GetName())
}

func TestDeleteNodes(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	id := "node4_test"

	ctx := context.Background()
	node, _ := jenkins.DeleteNode(ctx, id)
	assert.NotNil(t, node)
}

func TestCreateBuilds(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	jobs, _ := jenkins.GetAllJobs(ctx)
	for _, item := range jobs {
		queueID, _ = item.InvokeSimple(ctx, map[string]string{"params1": "param1"})
		item.Poll(ctx)
		isQueued, _ := item.IsQueued(ctx)
		assert.Equal(t, true, isQueued)
		time.Sleep(10 * time.Second)
		builds, _ := item.GetAllBuildIds(ctx)

		assert.True(t, (len(builds) > 0))

	}
}

func TestGetQueueItem(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	task, err := jenkins.GetQueueItem(ctx, queueID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Raw == nil || task.Raw.ID != queueID {
		t.Fatal()
	}
}

func TestParseBuildHistory(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	r, err := os.Open("_tests/build_history.txt")
	if err != nil {
		panic(err)
	}
	history := parseBuildHistory(r)
	assert.True(t, len(history) == 3)
}

func TestCreateViews(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	list_view, err := jenkins.CreateView(ctx, "test_list_view", LIST_VIEW)
	assert.Nil(t, err)
	assert.Equal(t, "test_list_view", list_view.GetName())
	assert.Equal(t, "", list_view.GetDescription())
	assert.Equal(t, 0, len(list_view.GetJobs()))

	my_view, err := jenkins.CreateView(ctx, "test_my_view", MY_VIEW)
	assert.Nil(t, err)
	assert.Equal(t, "test_my_view", my_view.GetName())
	assert.Equal(t, "", my_view.GetDescription())
	assert.Equal(t, 2, len(my_view.GetJobs()))

}

func TestGetAllJobs(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	jobs, _ := jenkins.GetAllJobs(ctx)
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, jobs[0].Raw.Color, "blue")
}

func TestGetAllNodes(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	nodes, _ := jenkins.GetAllNodes(ctx)
	assert.Equal(t, 3, len(nodes))
	assert.Equal(t, nodes[0].GetName(), "master")
}

func TestGetAllBuilds(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	builds, _ := jenkins.GetAllBuildIds(ctx, "Job1_test")
	for _, b := range builds {
		build, _ := jenkins.GetBuild(ctx, "Job1_test", b.Number)
		assert.Equal(t, "SUCCESS", build.GetResult())
	}
	assert.Equal(t, 1, len(builds))
}

func TestGetLabel(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	label, err := jenkins.GetLabel(ctx, "test_label")
	assert.Nil(t, err)
	assert.Equal(t, label.GetName(), "test_label")
	assert.Equal(t, 0, len(label.GetNodes()))

	label, err = jenkins.GetLabel(ctx, "jdk7")
	assert.Nil(t, err)
	assert.Equal(t, label.GetName(), "jdk7")
	assert.Equal(t, 1, len(label.GetNodes()))
	assert.Equal(t, "node3_test", label.GetNodes()[0].NodeName)

	//label, err = jenkins.GetLabel("jdk8")
	//assert.Nil(t, err)
	//assert.Equal(t, label.GetName(), "jdk8")
	//assert.Equal(t, 1, len(label.GetNodes()))
	//assert.Equal(t, "node2_test", label.GetNodes()[0].NodeName)
	//
	//label, err = jenkins.GetLabel("docker")
	//assert.Nil(t, err)
	//assert.Equal(t, label.GetName(), "docker")
	//assert.Equal(t, 1, len(label.GetNodes()))
	//assert.Equal(t, "node2_test", label.GetNodes()[0].NodeName)
}

func TestBuildMethods(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	build, _ := job.GetLastBuild(ctx)
	params := build.GetParameters()
	assert.Equal(t, "params1", params[0].Name)
}

func TestGetSingleJob(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	isRunning, _ := job.IsRunning(ctx)
	config, err := job.GetConfig(ctx)
	assert.Nil(t, err)
	assert.Equal(t, false, isRunning)
	assert.Contains(t, config, "<project>")
}

func TestEnableDisableJob(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	result, _ := job.Disable(ctx)
	assert.Equal(t, true, result)
	result, _ = job.Enable(ctx)
	assert.Equal(t, true, result)
}

func TestCopyDeleteJob(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	job, _ := jenkins.GetJob(ctx, "Job1_test")
	jobCopy, _ := job.Copy(ctx, "Job1_test_copy")
	assert.Equal(t, jobCopy.GetName(), "Job1_test_copy")
	jobDelete, _ := job.Delete(ctx)
	assert.Equal(t, true, jobDelete)
}

func TestGetPlugins(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	plugins, _ := jenkins.GetPlugins(ctx, 3)
	assert.Equal(t, 10, plugins.Count())
}

func TestGetViews(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	views, _ := jenkins.GetAllViews(ctx)
	assert.Equal(t, len(views), 3)
	assert.Equal(t, len(views[0].Raw.Jobs), 2)
}

func TestGetSingleView(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	view, _ := jenkins.GetView(ctx, "All")
	view2, _ := jenkins.GetView(ctx, "test_list_view")
	assert.Equal(t, len(view.Raw.Jobs), 2)
	assert.Equal(t, len(view2.Raw.Jobs), 0)
	assert.Equal(t, view2.Raw.Name, "test_list_view")
}

func TestCreateFolder(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	folder1ID := "folder1_test"
	folder2ID := "folder2_test"

	folder1, err := jenkins.CreateFolder(ctx, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, folder1ID, folder1.GetName())

	folder2, err := jenkins.CreateFolder(ctx, folder2ID, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, folder2ID, folder2.GetName())
}

func TestCreateJobInFolder(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	jobName := "Job_test"
	job_data := getFileAsString("job.xml")

	job1, err := jenkins.CreateJobInFolder(ctx, job_data, jobName, "folder1_test")
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, jobName, job1.GetName())

	job2, err := jenkins.CreateJobInFolder(ctx, job_data, jobName, "folder1_test", "folder2_test")
	assert.Nil(t, err)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, jobName, job2.GetName())
}

func TestGetFolder(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	folder1ID := "folder1_test"
	folder2ID := "folder2_test"

	folder1, err := jenkins.GetFolder(ctx, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, folder1ID, folder1.GetName())

	folder2, err := jenkins.GetFolder(ctx, folder2ID, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, folder2ID, folder2.GetName())
}
func TestInstallPlugin(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()

	err := jenkins.InstallPlugin(ctx, "packer", "1.4")

	assert.Nil(t, err, "Could not install plugin")
}

func TestConcurrentRequests(t *testing.T) {
	if _, ok := os.LookupEnv(integration_test); !ok {
		return
	}
	ctx := context.Background()
	for i := 0; i <= 16; i++ {
		go func() {
			jenkins.GetAllJobs(ctx)
			jenkins.GetAllViews(ctx)
			jenkins.GetAllNodes(ctx)
		}()
	}
}

func getFileAsString(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
