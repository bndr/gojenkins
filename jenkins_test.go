package gojenkins

import (
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
	jenkins = CreateJenkins(nil, "http://localhost:8080", "admin", "admin")
	_, err := jenkins.Init()
	assert.Nil(t, err, "Jenkins Initialization should not fail")
}

func TestCreateJobs(t *testing.T) {
	job1ID := "Job1_test"
	job2ID := "job2_test"
	job_data := getFileAsString("job.xml")

	job1, err := jenkins.CreateJob(job_data, job1ID)
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, job1ID, job1.GetName())

	job2, _ := jenkins.CreateJob(job_data, job2ID)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, job2ID, job2.GetName())
}


func TestCreateNodes(t *testing.T) {

	id1 := "node1_test"
	//id2 := "node2_test"
	id3 := "node3_test"
	id4 := "node4_test"

	jnlp := map[string]string{"method": "JNLPLauncher"}
	//ssh := map[string]string{"method": "SSHLauncher"}

	node1, _ := jenkins.CreateNode(id1, 1, "Node 1 Description", "/var/lib/jenkins", "", jnlp)
	assert.Equal(t, id1, node1.GetName())

	//node2, _ := jenkins.CreateNode(id2, 1, "Node 2 Description", "/var/lib/jenkins", "jdk8 docker", ssh)
	//assert.Equal(t, id2, node2.GetName())

	node3, _ := jenkins.CreateNode(id3, 1, "Node 3 Description", "/var/lib/jenkins", "jdk7")
	assert.Equal(t, id3, node3.GetName())
	node4, _ := jenkins.CreateNode(id4, 1, "Node 4 Description", "/var/lib/jenkins", "jdk7")
	assert.Equal(t, id4, node4.GetName())
}

func TestDeleteNodes(t *testing.T) {
	id := "node4_test"
	node, _ := jenkins.DeleteNode(id)
	assert.NotNil(t, node)
}

func TestCreateBuilds(t *testing.T) {
	jobs, _ := jenkins.GetAllJobs()
	for _, item := range jobs {
		queueID, _ = item.InvokeSimple(map[string]string{"params1": "param1"})
		item.Poll()
		isQueued, _ := item.IsQueued()
		assert.Equal(t, true, isQueued)
		time.Sleep(10 * time.Second)
		builds, _ := item.GetAllBuildIds()

		assert.True(t, (len(builds) > 0))

	}
}

func TestGetQueueItem(t *testing.T) {
	task, err := jenkins.GetQueueItem(queueID)
	if err != nil {
		t.Fatal(err)
	}
	if task.Raw == nil || task.Raw.ID != queueID {
		t.Fatal()
	}
}

func TestParseBuildHistory(t *testing.T) {
	r, err := os.Open("_tests/build_history.txt")
	if err != nil {
		panic(err)
	}
	history := parseBuildHistory(r)
	assert.True(t, len(history) == 3)
}

func TestCreateViews(t *testing.T) {
	list_view, err := jenkins.CreateView("test_list_view", LIST_VIEW)
	assert.Nil(t, err)
	assert.Equal(t, "test_list_view", list_view.GetName())
	assert.Equal(t, "", list_view.GetDescription())
	assert.Equal(t, 0, len(list_view.GetJobs()))

	my_view, err := jenkins.CreateView("test_my_view", MY_VIEW)
	assert.Nil(t, err)
	assert.Equal(t, "test_my_view", my_view.GetName())
	assert.Equal(t, "", my_view.GetDescription())
	assert.Equal(t, 2, len(my_view.GetJobs()))

}

func TestGetAllJobs(t *testing.T) {
	jobs, _ := jenkins.GetAllJobs()
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, jobs[0].Raw.Color, "blue")
}

func TestGetAllNodes(t *testing.T) {
	nodes, _ := jenkins.GetAllNodes()
	assert.Equal(t, 3, len(nodes))
	assert.Equal(t, nodes[0].GetName(), "master")
}

func TestGetAllBuilds(t *testing.T) {
	builds, _ := jenkins.GetAllBuildIds("Job1_test")
	for _, b := range builds {
		build, _ := jenkins.GetBuild("Job1_test", b.Number)
		assert.Equal(t, "SUCCESS", build.GetResult())
	}
	assert.Equal(t, 1, len(builds))
}

func TestGetLabel(t *testing.T) {
	label, err := jenkins.GetLabel("test_label")
	assert.Nil(t, err)
	assert.Equal(t, label.GetName(), "test_label")
	assert.Equal(t, 0, len(label.GetNodes()))

	label, err = jenkins.GetLabel("jdk7")
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
	job, _ := jenkins.GetJob("Job1_test")
	build, _ := job.GetLastBuild()
	params := build.GetParameters()
	assert.Equal(t, "params1", params[0].Name)
}

func TestGetSingleJob(t *testing.T) {
	job, _ := jenkins.GetJob("Job1_test")
	isRunning, _ := job.IsRunning()
	config, err := job.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, false, isRunning)
	assert.Contains(t, config, "<project>")
}

func TestEnableDisableJob(t *testing.T) {
	job, _ := jenkins.GetJob("Job1_test")
	result, _ := job.Disable()
	assert.Equal(t, true, result)
	result, _ = job.Enable()
	assert.Equal(t, true, result)
}

func TestCopyDeleteJob(t *testing.T) {
	job, _ := jenkins.GetJob("Job1_test")
	jobCopy, _ := job.Copy("Job1_test_copy")
	assert.Equal(t, jobCopy.GetName(), "Job1_test_copy")
	jobDelete, _ := job.Delete()
	assert.Equal(t, true, jobDelete)
}

func TestGetPlugins(t *testing.T) {
	plugins, _ := jenkins.GetPlugins(3)
	assert.Equal(t, 10, plugins.Count())
}

func TestGetViews(t *testing.T) {
	views, _ := jenkins.GetAllViews()
	assert.Equal(t, len(views), 3)
	assert.Equal(t, len(views[0].Raw.Jobs), 2)
}

func TestGetSingleView(t *testing.T) {
	view, _ := jenkins.GetView("All")
	view2, _ := jenkins.GetView("test_list_view")
	assert.Equal(t, len(view.Raw.Jobs), 2)
	assert.Equal(t, len(view2.Raw.Jobs), 0)
	assert.Equal(t, view2.Raw.Name, "test_list_view")
}

func TestCreateFolder(t *testing.T) {
	folder1ID := "folder1_test"
	folder2ID := "folder2_test"

	folder1, err := jenkins.CreateFolder(folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, folder1ID, folder1.GetName())

	folder2, err := jenkins.CreateFolder(folder2ID, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, folder2ID, folder2.GetName())
}

func TestCreateJobInFolder(t *testing.T) {
	jobName := "Job_test"
	job_data := getFileAsString("job.xml")

	job1, err := jenkins.CreateJobInFolder(job_data, jobName, "folder1_test")
	assert.Nil(t, err)
	assert.NotNil(t, job1)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, jobName, job1.GetName())

	job2, err := jenkins.CreateJobInFolder(job_data, jobName, "folder1_test", "folder2_test")
	assert.Nil(t, err)
	assert.NotNil(t, job2)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, jobName, job2.GetName())
}

func TestGetFolder(t *testing.T) {
	folder1ID := "folder1_test"
	folder2ID := "folder2_test"

	folder1, err := jenkins.GetFolder(folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder1)
	assert.Equal(t, folder1ID, folder1.GetName())

	folder2, err := jenkins.GetFolder(folder2ID, folder1ID)
	assert.Nil(t, err)
	assert.NotNil(t, folder2)
	assert.Equal(t, folder2ID, folder2.GetName())
}
func TestInstallPlugin(t *testing.T) {

	err := jenkins.InstallPlugin("packer", "1.4")

	assert.Nil(t, err, "Could not install plugin")
}

func TestConcurrentRequests(t *testing.T) {
	for i := 0; i <= 16; i++ {
		go func() {
			jenkins.GetAllJobs()
			jenkins.GetAllViews()
			jenkins.GetAllNodes()
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
