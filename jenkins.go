package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
)

type BasicAuth struct {
	Username string
	Password string
}

type TokenAuth struct {
	Username string
	Token    string
}

type Jenkins struct {
	Server    string
	Version   string
	Requester *Requester
}

// Jenkins

func (j *Jenkins) Init() *Jenkins {

	// Skip SSL Verification?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !j.Requester.SslVerify},
	}

	cookies, _ := cookiejar.New(nil)

	client := &http.Client{
		Transport: tr,
		Jar:       cookies,
	}
	if j.Requester.Client == nil {
		j.Requester.Client = client
	}

	// Check Connection
	resp := new(ExecutorResponse)
	raw := j.Requester.Do("GET", "/", nil, resp)
	j.Version = raw.Header.Get("X-Jenkins")
	if resp == nil {
		panic("Connection Failed, Please verify that the host and credentials are correct.")
	}
	return j
}

func (j *Jenkins) Info() {

}

func (j *Jenkins) CreateNode() {

}

func (j *Jenkins) CreateBuild() {

}

func (j *Jenkins) CreateJob() {

}

func (j *Jenkins) GetNode(id string) Node {
	node := Node{Raw: new(nodeResponse), Requester: j.Requester}
	j.Requester.Get("/computer/"+id, node.Raw)
	return node
}

func (j *Jenkins) GetBuild(job string, number string) Build {
	build := Build{Raw: new(buildResponse), Requester: j.Requester}
	j.Requester.Get("/job/"+job+"/"+number, build.Raw)
	return build
}

func (j *Jenkins) GetJob(id string) Job {
	job := Job{Raw: new(jobResponse), Requester: j.Requester}
	j.Requester.Get("/job/"+id, job.Raw)
	return job
}

func (j *Jenkins) GetAllNodes() []Node {
	computers := new(Computers)
	j.Requester.Get("/computer", computers)
	nodes := make([]Node, len(computers.Computers))
	for i, node := range computers.Computers {
		nodes[i] = Node{Raw: &node, Requester: j.Requester}
	}
	return nodes
}

func (j *Jenkins) GetAllBuilds(job string, options ...interface{}) []Build {
	jobObj := j.GetJob(job)
	builds := make([]Build, len(jobObj.Raw.Builds))
	preload := false
	if len(options) > 0 && options[0].(bool) {
		preload = true
	}
	for i, build := range jobObj.Raw.Builds {
		if preload == false {
			builds[i] = Build{Raw: &buildResponse{Number: build.Number, URL: build.URL}, Requester: j.Requester}
		} else {
			builds[i] = j.GetBuild(job, strconv.Itoa(build.Number))
		}
	}
	return builds
}

func (j *Jenkins) GetAllJobs(preload bool) []Job {
	exec := Executor{Raw: new(ExecutorResponse), Requester: j.Requester}
	j.Requester.Get("/", exec.Raw)
	jobs := make([]Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		if preload == false {
			jobs[i] = Job{Raw: &jobResponse{Name: job.Name, Color: job.Color, URL: job.URL}, Requester: j.Requester}
		} else {
			jobs[i] = j.GetJob(job.Name)
		}
	}
	return jobs
}

func CreateJenkins(base string, username string, password string) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: false, Headers: http.Header{}}
	j.Requester.BasicAuth = &BasicAuth{Username: username, Password: password}
	return j
}

func main() {
	j := CreateJenkins("http://localhost:8080/", "admin", "admin").Init()

	//fmt.Printf("%#v\n", j.GetJob("testJobName").Raw.Description)
	job := j.GetAllJobs(true)[0].GetName()
	fmt.Printf("%#v\n", job)
	fmt.Printf("%#v\n", j.GetNode("testNode"))

}
