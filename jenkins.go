// Copyright 2015 Vadim Kravcenko
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
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	Raw       *ExecutorResponse
	Requester *Requester
}

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Init Method. Should be called after creating a Jenkins Instance.
// e.g jenkins,err := CreateJenkins("url").Init()
// HTTP Client is set here, Connection to jenkins is tested here.
func (j *Jenkins) Init(ctx context.Context) (*Jenkins, error) {
	j.initLoggers()

	// Check Connection
	j.Raw = new(ExecutorResponse)
	rsp, err := j.Requester.GetJSON(ctx, "/", j.Raw, nil)
	if err != nil {
		return nil, err
	}

	j.Version = rsp.Header.Get("X-Jenkins")
	if j.Raw == nil || rsp.StatusCode != http.StatusOK {
		return nil, errors.New("Connection Failed, Please verify that the host and credentials are correct.")
	}

	return j, nil
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
func (j *Jenkins) Info(ctx context.Context) (*ExecutorResponse, error) {
	rsp, err := j.Requester.GetJSON(ctx, "/", j.Raw, nil)
	if err != nil {
		return nil, err
	}

	// The cached version which is set in Init(), might get staled, update
	j.Version = rsp.Header.Get("X-Jenkins")

	return j.Raw, nil
}

// SafeRestart jenkins, restart will be done when there are no jobs running
func (j *Jenkins) SafeRestart(ctx context.Context) error {
	_, err := j.Requester.Post(ctx, "/safeRestart", strings.NewReader(""), struct{}{}, map[string]string{})
	return err
}

// CreateNode Create a new Node
// Can be JNLPLauncher or SSHLauncher
// Example : jenkins.CreateNode("nodeName", 1, "Description", "/var/lib/jenkins", "jdk8 docker", map[string]string{"method": "JNLPLauncher"})
// By Default JNLPLauncher is created
// Multiple labels should be separated by blanks
func (j *Jenkins) CreateNode(ctx context.Context, name string, numExecutors int, description string, remoteFS string, label string, options ...interface{}) (*Node, error) {
	params := map[string]string{"method": "JNLPLauncher"}

	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}

	if _, ok := params["method"]; !ok {
		params["method"] = "JNLPLauncher"
	}

	method := params["method"]
	var launcher map[string]string
	switch method {
	case "":
		fallthrough
	case "JNLPLauncher":
		launcher = map[string]string{"stapler-class": "hudson.slaves.JNLPLauncher"}
	case "SSHLauncher":
		launcher = map[string]string{
			"stapler-class":        "hudson.plugins.sshslaves.SSHLauncher",
			"$class":               "hudson.plugins.sshslaves.SSHLauncher",
			"host":                 params["host"],
			"port":                 params["port"],
			"credentialsId":        params["credentialsId"],
			"jvmOptions":           params["jvmOptions"],
			"javaPath":             params["javaPath"],
			"prefixStartSlaveCmd":  params["prefixStartSlaveCmd"],
			"suffixStartSlaveCmd":  params["suffixStartSlaveCmd"],
			"maxNumRetries":        params["maxNumRetries"],
			"retryWaitTime":        params["retryWaitTime"],
			"lanuchTimeoutSeconds": params["lanuchTimeoutSeconds"],
			"type":                 "hudson.slaves.DumbSlave",
			"stapler-class-bag":    "true"}
	default:
		return nil, errors.New("launcher method not supported")
	}

	node := &Node{Jenkins: j, Raw: new(NodeResponse), Base: "/computer/" + name}
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
			"labelString":        label,
			"retentionsStrategy": map[string]string{"stapler-class": "hudson.slaves.RetentionStrategy$Always"},
			"nodeProperties":     map[string]string{"stapler-class-bag": "true"},
			"launcher":           launcher,
		}),
	}

	resp, err := j.Requester.Post(ctx, "/computer/doCreateItem", nil, nil, qr)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 400 {
		_, err := node.Poll(ctx)
		if err != nil {
			return nil, err
		}
		return node, nil
	}
	return nil, errors.New(strconv.Itoa(resp.StatusCode))
}

//DeleteNode Delete a Jenkins slave node
func (j *Jenkins) DeleteNode(ctx context.Context, name string) (bool, error) {
	node := Node{Jenkins: j, Raw: new(NodeResponse), Base: "/computer/" + name}
	return node.Delete(ctx)
}

//CreateFolder Create a new folder
// This folder can be nested in other parent folders
// Example: jenkins.CreateFolder("newFolder", "grandparentFolder", "parentFolder")
func (j *Jenkins) CreateFolder(ctx context.Context, name string, parents ...string) (*Folder, error) {
	folderObj := &Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, name), "/job/")}
	folder, err := folderObj.Create(ctx, name)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

//CreateJobInFolder Create a new job in the folder
// Example: jenkins.CreateJobInFolder("<config></config>", "newJobName", "myFolder", "parentFolder")
func (j *Jenkins) CreateJobInFolder(ctx context.Context, config string, jobName string, parentIDs ...string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, jobName), "/job/")}
	qr := map[string]string{
		"name": jobName,
	}
	job, err := jobObj.Create(ctx, config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

//ManageClouds Manage Jenkins Clouds
func (j *Jenkins) ManageClouds(ctx context.Context, cg CloudConfig) (*KubernetesCloud, error) {
	objK8s := &KubernetesCloud{Jenkins: j, Raw: new([]CloudResponse), K8sCloud: cg, Base: "/scriptText"}
	k8s, err := objK8s.CloudConfigure(ctx)
	if err != nil {
		return nil, err
	}
	return k8s, nil
}

//CreateJob Create a new job from config File
// Method takes XML string as first parameter, and if the name is not specified in the config file
// takes name as string as second parameter
// e.g jenkins.CreateJob("<config></config>","newJobName")
func (j *Jenkins) CreateJob(ctx context.Context, config string, options ...interface{}) (*Job, error) {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	} else {
		return nil, errors.New("Error Creating Job, job name is missing")
	}
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + qr["name"]}
	job, err := jobObj.Create(ctx, config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

//UpdateJob Update a job.
// If a job is exist, update its config
func (j *Jenkins) UpdateJob(ctx context.Context, job string, config string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	jobObj.UpdateConfig(ctx, config)
	return &jobObj
}

//RenameJob First parameter job old name, Second parameter job new name.
func (j *Jenkins) RenameJob(ctx context.Context, job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	jobObj.Rename(ctx, name)
	return &jobObj
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
func (j *Jenkins) CopyJob(ctx context.Context, copyFrom string, newName string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + copyFrom}
	_, err := job.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return job.Copy(ctx, newName)
}

// Delete a job.
func (j *Jenkins) DeleteJob(ctx context.Context, name string) (bool, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
	return job.Delete(ctx)
}

// Get a job object
func (j *Jenkins) GetJobObj(ctx context.Context, name string) *Job {
	return &Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
}

// Invoke a job.
// First parameter job name, second parameter is optional Build parameters.
// Returns queue id
func (j *Jenkins) BuildJob(ctx context.Context, name string, params map[string]string) (int64, error) {
	job := j.GetJobObj(ctx, name)
	return job.InvokeSimple(ctx, params)
}

// A task in queue will be assigned a build number in a job after a few seconds.
// this function will return the build object.
func (j *Jenkins) GetBuildFromQueueID(ctx context.Context, job *Job, queueid int64) (*Build, error) {
	task, err := j.GetQueueItem(ctx, queueid)
	if err != nil {
		return nil, err
	}
	// Jenkins queue API has about 4.7second quiet period
	for task.Raw.Executable.Number == 0 {
		time.Sleep(1000 * time.Millisecond)
		_, err = task.Poll(ctx)
		if err != nil {
			return nil, err
		}
	}

	build, err := job.GetBuild(ctx, task.Raw.Executable.Number)
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (j *Jenkins) GetNode(ctx context.Context, name string) (*Node, error) {
	node := Node{Jenkins: j, Raw: new(NodeResponse), Base: "/computer/" + name}
	status, err := node.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &node, nil
	}
	return nil, errors.New("No node found")
}

func (j *Jenkins) GetLabel(ctx context.Context, name string) (*Label, error) {
	label := Label{Jenkins: j, Raw: new(LabelResponse), Base: "/label/" + name}
	status, err := label.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &label, nil
	}
	return nil, errors.New("No label found")
}

func (j *Jenkins) GetBuild(ctx context.Context, jobName string, number int64) (*Build, error) {
	job, err := j.GetJob(ctx, jobName)
	if err != nil {
		return nil, err
	}
	build, err := job.GetBuild(ctx, number)

	if err != nil {
		return nil, err
	}
	return build, nil
}

func (j *Jenkins) GetJob(ctx context.Context, id string, parentIDs ...string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, id), "/job/")}
	status, err := job.Poll(ctx)
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetSubJob(ctx context.Context, parentId string, childId string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + parentId + "/job/" + childId}
	status, err := job.Poll(ctx)
	if err != nil {
		return nil, fmt.Errorf("trouble polling job: %v", err)
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetFolder(ctx context.Context, id string, parents ...string) (*Folder, error) {
	folder := Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, id), "/job/")}
	status, err := folder.Poll(ctx)
	if err != nil {
		return nil, fmt.Errorf("trouble polling folder: %v", err)
	}
	if status == 200 {
		return &folder, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetAllNodes(ctx context.Context) ([]*Node, error) {
	computers := new(Computers)

	qr := map[string]string{
		"depth": "1",
	}

	_, err := j.Requester.GetJSON(ctx, "/computer", computers, qr)
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, len(computers.Computers))
	for i, node := range computers.Computers {
		nodes[i] = &Node{Jenkins: j, Raw: node, Base: "/computer/" + node.DisplayName}
	}

	return nodes, nil
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use jenkins.GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)
func (j *Jenkins) GetAllBuildIds(ctx context.Context, job string) ([]JobBuild, error) {
	jobObj, err := j.GetJob(ctx, job)
	if err != nil {
		return nil, err
	}
	return jobObj.GetAllBuildIds(ctx)
}

// Get Only Array of Job Names, Color, URL
// Does not query each single Job.
func (j *Jenkins) GetAllJobNames(ctx context.Context) ([]InnerJob, error) {
	exec := Executor{Raw: new(ExecutorResponse), Jenkins: j}
	_, err := j.Requester.GetJSON(ctx, "/", exec.Raw, nil)

	if err != nil {
		return nil, err
	}

	return exec.Raw.Jobs, nil
}

// Get All Possible Job Objects.
// Each job will be queried.
func (j *Jenkins) GetAllJobs(ctx context.Context) ([]*Job, error) {
	exec := Executor{Raw: new(ExecutorResponse), Jenkins: j}
	_, err := j.Requester.GetJSON(ctx, "/", exec.Raw, nil)

	if err != nil {
		return nil, err
	}

	jobs := make([]*Job, len(exec.Raw.Jobs))
	for i, job := range exec.Raw.Jobs {
		ji, err := j.GetJob(ctx, job.Name)
		if err != nil {
			return nil, err
		}
		jobs[i] = ji
	}
	return jobs, nil
}

// Returns a Queue
func (j *Jenkins) GetQueue(ctx context.Context) (*Queue, error) {
	q := &Queue{Jenkins: j, Raw: new(queueResponse), Base: j.GetQueueUrl()}
	_, err := q.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (j *Jenkins) GetQueueUrl() string {
	return "/queue"
}

// GetQueueItem returns a single queue Task
func (j *Jenkins) GetQueueItem(ctx context.Context, id int64) (*Task, error) {
	t := &Task{Raw: new(taskResponse), Jenkins: j, Base: j.getQueueItemURL(id)}
	_, err := t.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (j *Jenkins) getQueueItemURL(id int64) string {
	return fmt.Sprintf("/queue/item/%d", id)
}

// Get Artifact data by Hash
func (j *Jenkins) GetArtifactData(ctx context.Context, id string) (*FingerPrintResponse, error) {
	fp := FingerPrint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(FingerPrintResponse)}
	return fp.GetInfo(ctx)
}

//GetPlugins Returns the list of all plugins installed on the Jenkins server.
// You can supply depth parameter, to limit how much data is returned.
func (j *Jenkins) GetPlugins(ctx context.Context, depth int) (*Plugins, error) {
	p := Plugins{Jenkins: j, Raw: new(PluginResponse), Base: "/pluginManager", Depth: depth}
	_, err := p.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UninstallPlugin plugin otherwise returns error
func (j *Jenkins) UninstallPlugin(ctx context.Context, name string) error {
	url := fmt.Sprintf("/pluginManager/plugin/%s/doUninstall", name)
	resp, err := j.Requester.Post(ctx, url, strings.NewReader(""), struct{}{}, map[string]string{})
	if resp.StatusCode != 200 {
		return fmt.Errorf("Invalid status code returned: %d", resp.StatusCode)
	}
	return err
}

//HasPlugin Check if the plugin is installed on the server.
// Depth level 1 is used. If you need to go deeper, you can use GetPlugins, and iterate through them.
func (j *Jenkins) HasPlugin(ctx context.Context, name string) (*Plugin, error) {
	p, err := j.GetPlugins(ctx, 1)

	if err != nil {
		return nil, err
	}
	return p.Contains(name), nil
}

//InstallPlugin with given name
func (j *Jenkins) InstallPlugin(ctx context.Context, name string) (*Plugins, error) {
	pluginObj := Plugins{Jenkins: j, Raw: new(PluginResponse), Base: "/pluginManager", Depth: 1}
	plugins, err := pluginObj.Install(ctx, name)
	if err != nil {
		return nil, err
	}
	return plugins, nil
}

// Verify FingerPrint
func (j *Jenkins) ValidateFingerPrint(ctx context.Context, id string) (bool, error) {
	fp := FingerPrint{Jenkins: j, Base: "/fingerprint/", Id: id, Raw: new(FingerPrintResponse)}
	valid, err := fp.Valid(ctx)
	if err != nil {
		return false, err
	}
	if valid {
		return true, nil
	}
	return false, nil
}

func (j *Jenkins) GetView(ctx context.Context, name string) (*View, error) {
	url := "/view/" + name
	view := View{Jenkins: j, Raw: new(ViewResponse), Base: url}
	_, err := view.Poll(ctx)
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (j *Jenkins) GetAllViews(ctx context.Context) ([]*View, error) {
	_, err := j.Poll(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]*View, len(j.Raw.Views))
	for i, v := range j.Raw.Views {
		views[i], _ = j.GetView(ctx, v.Name)
	}
	return views, nil
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
func (j *Jenkins) CreateView(ctx context.Context, name string, viewType string) (*View, error) {
	view := &View{Jenkins: j, Raw: new(ViewResponse), Base: "/view/" + name}
	endpoint := "/createView"
	data := map[string]string{
		"name":   name,
		"mode":   viewType,
		"Submit": "OK",
		"json": makeJson(map[string]string{
			"name": name,
			"mode": viewType,
		}),
	}
	r, err := j.Requester.Post(ctx, endpoint, nil, view.Raw, data)

	if err != nil {
		return nil, err
	}

	if r.StatusCode == 200 {
		return j.GetView(ctx, name)
	}
	return nil, errors.New(strconv.Itoa(r.StatusCode))
}

func (j *Jenkins) Poll(ctx context.Context) (int, error) {
	resp, err := j.Requester.GetJSON(ctx, "/", j.Raw, nil)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

// Creates a new Jenkins Instance
// Optional parameters are: client, username, password or token
// After creating an instance call init method.
func CreateJenkins(client *http.Client, base string, auth ...interface{}) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: true, Client: client}
	if j.Requester.Client == nil {
		j.Requester.Client = http.DefaultClient
	}
	if len(auth) == 2 {
		j.Requester.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}
