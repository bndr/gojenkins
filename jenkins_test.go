package gojenkins

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"math/rand"
	"time"
)

var (
	jenkins *Jenkins
)


func init() {
	jenkins = CreateJenkins("http://192.168.99.100:8080", "admin", "admin")
	jenkins.Init()
}

func TestCreateJobs(t *testing.T) {
	job1ID := "Job1_test"
	job2ID := "Job2_test"

	job1, _  := jenkins.CreateJob(readJson("job.xml"), job1ID)
	assert.Equal(t, "Some Job Description", job1.GetDescription())
	assert.Equal(t, job1ID, job1.GetName())


	job2, _  :=  jenkins.CreateJob(readJson("job.xml"), job2ID)
	assert.Equal(t, "Some Job Description", job2.GetDescription())
	assert.Equal(t, job2ID, job2.GetName())
}

func TestCreateNodes(t *testing.T) {

	id1 := "node1_test"
	id2 := "node2_test"
	node1, _  := jenkins.CreateNode(id1, 1, "Node 1 Description", "/var/lib/jenkins")

	assert.Equal(t, id1, node1.GetName())

	node2, _ := jenkins.CreateNode(id2, 1, "Node 2 Description", "/var/lib/jenkins")
	assert.Equal(t, id2, node2.GetName())
}

func TestCreateBuilds(t *testing.T) {
	jobs, _ := jenkins.GetAllJobs()
	for _, item := range jobs {
		item.InvokeSimple(map[string]string{"param1":"param1"})
		item.Poll()
		isQueued, _ := item.IsQueued()
		assert.Equal(t, true, isQueued)

		time.Sleep(10 * time.Second)
		builds, _ := item.GetAllBuildIds()

		assert.True(t, (len(builds) > 0))

	}
}

func TestCreateViews(t *testing.T) {
	resp, err := jenkins.CreateView("test_view", LIST_VIEW)
	fmt.Printf("%#v", err)
	fmt.Printf("%#v", resp)
}

func TestGetAllJobs(t *testing.T) {
	jobs, _ := jenkins.GetAllJobs()
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, jobs[0].Raw.Color, "blue")
}

func TestGetAllNodes(t *testing.T) {
	nodes := jenkins.GetAllNodes()
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

func TestBuildMethods(t *testing.T) {
	job, _ := jenkins.GetJob("Job1_test")
	build, _ := job.GetLastBuild()
	params := build.GetParameters()
	assert.Equal(t, "params1", params[0].Name)
}

func TestGetSingleJob(t *testing.T) {
	job, _ := jenkins.GetJob("Job1_test")
	isRunning, _ := job.IsRunning()
	assert.Equal(t, false, isRunning)
	assert.Contains(t, job.GetConfig(), "<project>")
}

func TestGetPlugins(t *testing.T) {
	plugins := jenkins.GetPlugins(3)
	assert.Equal(t, 19, plugins.Count())
}

func TestGetViews(t *testing.T) {
	views := jenkins.GetAllViews()
	assert.Equal(t, len(views), 2)
	assert.Equal(t, len(views[0].Raw.Jobs), 2)
}

func TestGetSingleView(t *testing.T) {
	view := jenkins.GetView("All")
	view2 := jenkins.GetView("test_view")
	assert.Equal(t, len(view.Raw.Jobs), 2)
	assert.Equal(t, len(view2.Raw.Jobs), 0)
	assert.Equal(t, view2.Raw.Name, "test_view")
}

func readJson(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}

func getRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}
