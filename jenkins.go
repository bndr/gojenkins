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
	Raw       *executorResponse
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
	j.Raw = new(executorResponse)
	j.Requester.Do("GET", "/", nil, j.Raw, nil)
	j.Version = j.Requester.LastResponse.Header.Get("X-Jenkins")
	if j.Raw == nil {
		panic("Connection Failed, Please verify that the host and credentials are correct.")
	}
	return j
}

func (j *Jenkins) Info() *executorResponse {
	j.Requester.Do("GET", "/", nil, j.Raw, nil)
	return j.Raw
}

func (j *Jenkins) CreateNode(name string, numExecutors int, description string, remoteFS string, options ...interface{}) *Node {
	node := j.GetNode(name)
	if node != nil {
		return node
	}
	node = &Node{Jenkins: j, Raw: new(nodeResponse), Requester: j.Requester, Base: "/computer/" + name}
	NODE_TYPE := "hudson.slaves.DumbSlave$DescriptorImpl"
	MODE := "NORMAL"
	qr := map[string]string{
		"name": name,
		"type": NODE_TYPE,
		"json": makeJson(map[string]interface{}{
			"name":               name,
			"nodeDescription":    description,
			"remoteFS":           remoteFS,
			"numExecutors":       numExecutors,
			"mode":               MODE,
			"type":               NODE_TYPE,
			"retentionsStrategy": map[string]string{"stapler-class": "hudson.slaves.RetentionStrategy$Always"},
			"nodeProperties":     map[string]string{"stapler-class-bag": "true"},
			"launcher":           map[string]string{"stapler-class": "hudson.slaves.JNLPLauncher"},
		}),
	}

	resp := node.Requester.GetXML("/computer/doCreateItem", nil, qr)
	if resp.StatusCode < 400 {
		node.Poll()
		return node
	}
	return nil
}

func (j *Jenkins) CreateJob(config string) *Job {
	job := Job{Jenkins: j, Raw: new(jobResponse), Requester: j.Requester}
	job.Create(config)
	return &job
}

func (j *Jenkins) RenameJob(job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(jobResponse), Requester: j.Requester, Base: "/job/" + job}
	jobObj.Rename(name)
	return &jobObj
}

func (j *Jenkins) CopyJob(copyFrom string, newName string) *Job {
	job := Job{Jenkins: j, Raw: new(jobResponse), Requester: j.Requester, Base: "/job/" + newName}
	return job.Copy(copyFrom, newName)
}

func (j *Jenkins) DeleteJob(name string) bool {
	job := Job{Jenkins: j, Raw: new(jobResponse), Requester: j.Requester, Base: "/job/" + name}
	return job.Delete()
}

func (j *Jenkins) BuildJob(name string, options ...interface{}) bool {
	job := Job{Jenkins: j, Raw: new(jobResponse), Requester: j.Requester, Base: "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.Invoke(nil, params)
}

func (j *Jenkins) GetNode(name string) *Node {
	node := Node{Jenkins: j, Raw: new(nodeResponse), Requester: j.Requester, Base: "/computers/" + name}
	if node.Poll() == 200 {
		return &node
	}
	return nil
}

func (j *Jenkins) GetBuild(job string, number string) *Build {
	build := Build{Jenkins: j, Raw: new(buildResponse), Depth: 1, Requester: j.Requester, Base: "/job/" + job + "/" + number}
	if build.Poll() == 200 {
		return &build
	}
	return nil
}

func (j *Jenkins) GetJob(id string) *Job {
	job := Job{Jenkins: j, Raw: new(jobResponse), Requester: j.Requester, Base: "/job/" + id}
	if job.Poll() == 200 {
		return &job
	}
	return nil
}

func (j *Jenkins) GetAllNodes() []*Node {
	computers := new(Computers)
	j.Requester.Get("/computer", computers, nil)
	nodes := make([]*Node, len(computers.Computers))
	for i, node := range computers.Computers {
		nodes[i] = &Node{Raw: &node, Requester: j.Requester}
	}
	return nodes
}

func (j *Jenkins) GetAllBuilds(job string, options ...interface{}) []*Build {
	jobObj := j.GetJob(job)
	builds := make([]*Build, len(jobObj.Raw.Builds))
	preload := false
	if len(options) > 0 && options[0].(bool) {
		preload = true
	}
	for i, build := range jobObj.Raw.Builds {
		if preload == false {
			builds[i] = &Build{
				Jenkins:   j,
				Depth:     1,
				Raw:       &buildResponse{Number: build.Number, URL: build.URL},
				Requester: j.Requester,
				Base:      "/job/" + jobObj.GetName() + "/" + string(build.Number)}
		} else {
			builds[i] = j.GetBuild(job, strconv.Itoa(build.Number))
		}
	}
	return builds
}

func (j *Jenkins) GetAllJobs(preload bool) []*Job {
	exec := executor{Raw: new(executorResponse), Requester: j.Requester}
	j.Requester.Get("/", exec.Raw, nil)
	jobs := make([]*Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		if preload == false {
			jobs[i] = &Job{
				Jenkins: j,
				Raw: &jobResponse{Name: job.Name,
					Color: job.Color,
					URL:   job.URL},
				Requester: j.Requester,
				Base:      "/job/" + job.Name}
		} else {
			jobs[i] = j.GetJob(job.Name)
		}
	}
	return jobs
}

func CreateJenkins(base string, auth ...interface{}) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: false, Headers: http.Header{}}
	if len(auth) == 2 {
		j.Requester.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}

func main() {
	j := CreateJenkins("http://localhost:8080/", "admin", "admin").Init()

	//fmt.Printf("%#v\n", j.GetJob("testJobName").Raw.Description)
	job := j.GetJob("testjib")
	ts := job.GetLastBuild()
	fmt.Println(ts.GetRevision())
	fmt.Println(ts.GetCauses())
	//fmt.Printf("%#v\n", j.GetJob("newjobsbb").Delete())
	//	fmt.Printf("%#v", j.CreateNode("wat23s1131sssasd1121", 2, "description", "/f/vs/sa/"))
}
