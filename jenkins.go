// Copyright 2014 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Gojenkins is a Jenkins Client in Go, that exposes the jenkins REST api in a more developer friendly way.
package gojenkins

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	Raw       *executorResponse
	Requester *Requester
}

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Init Method. Should be called after creating a Jenkins Instance.
// e.g jenkins := CreateJenkins("url").Init()
// HTTP Client is set here, Connection to jenkins is tested here.
func (j *Jenkins) Init() *Jenkins {
	j.initLoggers()
	// Skip SSL Verification?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !j.Requester.SslVerify},
	}

	if j.Requester.Client == nil {
		cookies, _ := cookiejar.New(nil)

		client := &http.Client{
			Transport: tr,
			Jar:       cookies,
		}
		j.Requester.Client = client
	}

	// Check Connection
	j.Raw = new(executorResponse)
	j.Requester.GetJSON("/", j.Raw, nil)
	j.Version = j.Requester.LastResponse.Header.Get("X-Jenkins")
	if j.Raw == nil {
		panic("Connection Failed, Please verify that the host and credentials are correct.")
	}
	return j
}

func (j *Jenkins) initLoggers() {
	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Get Basic Information About Jenkins
func (j *Jenkins) Info() *executorResponse {
	j.Requester.Do("GET", "/", nil, j.Raw, nil)
	return j.Raw
}

// Create a new Node
func (j *Jenkins) CreateNode(name string, numExecutors int, description string, remoteFS string, options ...interface{}) *Node {
	node := j.GetNode(name)
	if node != nil {
		return node
	}
	node = &Node{Jenkins: j, Raw: new(nodeResponse), Base: "/computer/" + name}
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

	resp := j.Requester.GetXML("/computer/doCreateItem", nil, qr)
	if resp.StatusCode < 400 {
		node.Poll()
		return node
	}
	return nil
}

// Create a new job from config File
// Method takes XML string as first parameter, and if the name is not specified in the config file
// takes name as string as second parameter
// e.g jenkins.CreateJob("<config></config>","newJobName")
func (j *Jenkins) CreateJob(config string, options ...interface{}) *Job {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	}
	job := Job{Jenkins: j, Raw: new(jobResponse)}
	job.Create(config, qr)
	return &job
}

// Rename a job.
// First parameter job old name, Second parameter job new name.
func (j *Jenkins) RenameJob(job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + job}
	jobObj.Rename(name)
	return &jobObj
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
func (j *Jenkins) CopyJob(copyFrom string, newName string) *Job {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + newName}
	return job.Copy(copyFrom, newName)
}

// Delete a job.
func (j *Jenkins) DeleteJob(name string) bool {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + name}
	return job.Delete()
}

// Invoke a job.
// First parameter job name, second parameter is optional Build parameters.
func (j *Jenkins) BuildJob(name string, options ...interface{}) bool {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.InvokeSimple(params)
}

func (j *Jenkins) GetNode(name string) *Node {
	node := Node{Jenkins: j, Raw: new(nodeResponse), Base: "/computer/" + name}
	if node.Poll() == 200 {
		return &node
	}
	return nil
}

func (j *Jenkins) GetBuild(jobName string, number int64) *Build {
	job := j.GetJob(jobName)
	if job != nil {
		return job.GetBuild(number)
	}
	return nil
}

func (j *Jenkins) GetJob(id string) *Job {
	job := Job{Jenkins: j, Raw: new(jobResponse), Base: "/job/" + id}
	if job.Poll() == 200 {
		return &job
	}
	return nil
}

func (j *Jenkins) GetAllNodes() []*Node {
	computers := new(Computers)
	j.Requester.GetJSON("/computer", computers, nil)
	nodes := make([]*Node, len(computers.Computers))
	for i, node := range computers.Computers {
		name := node.DisplayName
		// Special Case - Master Node
		if name == "master" {
			name = "(master)"
		}
		nodes[i] = &Node{Raw: &node, Jenkins: j, Base: "/computer/" + name}
		nodes[i].Poll()
	}
	return nodes
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use jenkins.GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)
func (j *Jenkins) GetAllBuildIds(job string) []jobBuild {
	jobObj := j.GetJob(job)
	if jobObj != nil {
		return jobObj.GetAllBuildIds()
	}
	return nil
}

// Get All Possible jobs
// If preload is true, the Job object will be preloaded before returning.
func (j *Jenkins) GetAllJobs(preload bool) []*Job {
	exec := Executor{Raw: new(executorResponse), Jenkins: j}
	j.Requester.GetJSON("/", exec.Raw, nil)
	jobs := make([]*Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		if preload == false {
			jobs[i] = &Job{
				Jenkins: j,
				Raw: &jobResponse{Name: job.Name,
					Color: job.Color,
					URL:   job.URL},
				Base: "/job/" + job.Name}
		} else {
			jobs[i] = j.GetJob(job.Name)
		}
	}
	return jobs
}

// Returns a Queue
func (j *Jenkins) GetQueue() *Queue {
	q := &Queue{Jenkins: j, Raw: new(queueResponse), Base: j.GetQueueUrl()}
	q.Poll()
	return q
}

func (j *Jenkins) GetQueueUrl() string {
	return "/queue"
}

// Get Artifact data by Hash
func (j *Jenkins) GetArtifactData(id string) *fingerPrintResponse {
	fp := Fingerprint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(fingerPrintResponse)}
	return fp.GetInfo()
}

// Returns the list of all plugins installed on the Jenkins server.
// You can supply depth parameter, to limit how much data is returned.
func (j *Jenkins) GetPlugins(depth int) *Plugins {
	p := Plugins{Jenkins: j, Raw: new(pluginResponse), Base: "/pluginManager", Depth: depth}
	p.Poll()
	return &p
}

// Check if the plugin is installed on the server.
// Depth level 1 is used. If you need to go deeper, you can use GetPlugins, and iterate through them.
func (j *Jenkins) HasPlugin(name string) *Plugin {
	p := j.GetPlugins(1)
	return p.Contains(name)
}

// Verify Fingerprint
func (j *Jenkins) ValidateFingerPrint(id string) bool {
	fp := Fingerprint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(fingerPrintResponse)}
	if fp.Valid() {
		Info.Printf("Jenkins says %s is valid", id)
		return true
	}
	return false
}

func (j *Jenkins) GetView(name string) *View {
	url := "/view/" + name
	view := View{Jenkins: j, Raw: new(viewResponse), Base: url}
	view.Poll()
	return &view
}

func (j *Jenkins) GetAllViews() []*View {
	j.Poll()
	views := make([]*View, len(j.Raw.Views))
	for i, v := range j.Raw.Views {
		views[i] = j.GetView(v.Name)
	}
	return views
}

// Create View
// First Parameter - name of the View
// Second parameter - Type
// Possible Types:
// 		gojenkins.LIST_VIEW
// 		gojenkins.NESTED_VIEW
// 		gojenkins.MY_VIEW
// 		gojenkins.DASHBOARD_VIEW
// 		gojenkins.PIPELINE_VIEW
// Example: jenkins.CreateView("newView",gojenkins.LIST_VIEW)

func (j *Jenkins) CreateView(name string, viewType string) bool {
	exists := j.GetView(name)
	if exists != nil {
		Error.Println("View Already exists.")
		return false
	}
	view := View{Jenkins: j, Raw: new(viewResponse), Base: "/view/" + name}
	url := "/createView"
	data := map[string]string{
		"name":   name,
		"type":   viewType,
		"Submit": "OK",
		"json": makeJson(map[string]string{
			"name": name,
			"mode": viewType,
		}),
	}
	r := j.Requester.Post(url, nil, view.Raw, data)
	return r.StatusCode == 200
}

func (j *Jenkins) Poll() int {
	j.Requester.GetJSON("/", j.Raw, nil)
	return j.Requester.LastResponse.StatusCode
}

// Creates a new Jenkins Instance
// Optional parameters are: username, password
// After creating an instance call init method.
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
