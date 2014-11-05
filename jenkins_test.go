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
	"/":                           func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("main.json")) },
	"/api/json":                   func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("main.json")) },
	"/job/testJob/api/json":       func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1.json")) },
	"/job/testJob/1/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job1_build1.json")) },
	"/jobtestJob2/api/json":       func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("job2.json")) },
	"/queue/api/json":             func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("queue.json")) },
	"/computer/api/json":          func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("nodes.json")) },
	"/computer/(master)/api/json": func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("node1.json")) },
	"/pluginManager/api/json":     func(rw http.ResponseWriter, req *http.Request) { fmt.Fprintln(rw, readJson("plugins.json")) },
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
	jobs := jenkins.GetAllJobs(true)
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, jobs[0].Raw.Color, "red")
}

func TestGetAllNodes(t *testing.T) {
	nodes := jenkins.GetAllNodes()
	assert.Equal(t, 2, len(nodes))
	assert.Equal(t, nodes[0].GetName(), "node2")
}

func TestGetAllBuilds(t *testing.T) {
	builds := jenkins.GetAllBuilds("testJob", true)
	assert.Equal(t, 2, len(builds))
	assert.Equal(t, "FAILURE", builds[0].GetResult())
	assert.Equal(t, "FAILURE", builds[0].GetResult())
}

func TestGetSingleJob(t *testing.T) {
	job := jenkins.GetJob("testJob")
	assert.Equal(t, false, job.IsRunning())
}

func TestGetPlugins(t *testing.T) {
	plugins := jenkins.GetPlugins(3)
	assert.Equal(t, plugins.Count(), 23)
}

func TestGetNodes(t *testing.T) {

}

func TestGetSingleNode(t *testing.T) {

}

func readJson(path string) string {
	buf, err := ioutil.ReadFile("_tests/" + path)
	if err != nil {
		panic(err)
	}

	return string(buf)
}
