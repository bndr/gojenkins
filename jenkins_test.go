package gojenkins

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testMux *http.ServeMux
	testSrv *httptest.Server
	jenkins *Jenkins
)

var paths = map[string]func(http.ResponseWriter, *http.Request){
	"/":         func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("main.json")) },
	"/api/json": func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("main.json")) },
	"/job/testJob/api/json": func(rw http.ResponseWriter, req *http.Request) {
		if req.FormValue("tree") != "" {
			fmt.Fprintln(rw, readJson("allBuilds.json"))
		} else {
			fmt.Fprintln(rw, readJson("job1.json"))
		}
	},
	"/job/testJob/config.xml/api/json": func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			fmt.Fprintln(rw, readJson("job.xml"))
		} else {
			rw.WriteHeader(400)
			fmt.Fprintln(rw, "Bad request")
		}
	},
	"/job/testJob/1/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/job/testJob/2/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/job/testJob/3/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/job/testJob/4/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/job/testJob/5/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/job/testJob/6/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/job/testJob2/api/json":      func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job2.json")) },
	"/queue/api/json":             func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("queue.json")) },
	"/computer/api/json":          func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("nodes.json")) },
	"/computer/(master)/api/json": func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("node1.json")) },
	"/pluginManager/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("plugins.json")) },
	"/view/test/api/json":         func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("view1.json")) },
}

func init() {
	testMux = http.NewServeMux()
	testSrv = httptest.NewServer(testMux)
	jenkins = CreateJenkins(testSrv.URL)
	for route, f := range paths {
		testMux.HandleFunc(route, f)
	}
	jenkins.Init()
}

func TestGetAllJobs(t *testing.T) {
	jobs, _ := jenkins.GetAllJobs()
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, jobs[0].Raw.Color, "red")
}

func TestGetAllNodes(t *testing.T) {
	nodes := jenkins.GetAllNodes()
	assert.Equal(t, 2, len(nodes))
	assert.Equal(t, nodes[0].GetName(), "master")
}

func TestGetAllBuilds(t *testing.T) {
	builds, _ := jenkins.GetAllBuildIds("testJob")
	for _, b := range builds {
		build, _ := jenkins.GetBuild("testJob", b.Number)
		assert.Equal(t, "FAILURE", build.GetResult())
		assert.Equal(t, "FAILURE", build.GetResult())
	}
	assert.Equal(t, 6, len(builds))
}

func TestBuildMethods(t *testing.T) {
	job, _ := jenkins.GetJob("testJob")
	build, _ := job.GetLastBuild()
	params := build.GetParameters()
	assert.Equal(t, "param1", params[0].Name)
	// TODO: All Methods
}

func TestGetSingleJob(t *testing.T) {
	job, _ := jenkins.GetJob("testJob")
	isRunning, _ := job.IsRunning()
	assert.Equal(t, false, isRunning)
	assert.Contains(t, job.GetConfig(), "<project>")
	// TODO: All Methods
}

func TestGetPlugins(t *testing.T) {
	plugins := jenkins.GetPlugins(3)
	assert.Equal(t, plugins.Count(), 23)
}

func TestGetViews(t *testing.T) {
	views := jenkins.GetAllViews()
	assert.Equal(t, len(views), 2)
	assert.Equal(t, len(views[1].Raw.Jobs), 1)
}

func TestGetSingleView(t *testing.T) {
	view := jenkins.GetView("test")
	assert.Equal(t, len(view.Raw.Jobs), 1)
	assert.Equal(t, view.Raw.Name, "test")
}

func TestCreation(t *testing.T) {
	// TODO
}

func TestDeletion(t *testing.T) {
	// TODO
}

func readJson(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
