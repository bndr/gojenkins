# Jenkins API Client for Go

[![GoDoc](https://godoc.org/github.com/bndr/gojenkins?status.svg)](https://godoc.org/github.com/bndr/gojenkins)
[![Go Report Cart](https://goreportcard.com/badge/github.com/bndr/gojenkins)](https://goreportcard.com/report/github.com/bndr/gojenkins)
[![Build Status](https://travis-ci.org/bndr/gojenkins.svg?branch=master)](https://travis-ci.org/bndr/gojenkins)

## About

Jenkins is the most popular Open Source Continuous Integration system. This Library will help you interact with Jenkins in a more developer-friendly way.

These are some of the features that are currently implemented:

* Get information on test-results of completed/failed build
* Ability to query Nodes, and manipulate them. Start, Stop, set Offline.
* Ability to query Jobs, and manipulate them.
* Get Plugins, Builds, Artifacts, Fingerprints
* Validate Fingerprints of Artifacts
* Get Current Queue, Cancel Tasks
* etc. For all methods go to GoDoc Reference.

## Installation

    go get github.com/bndr/gojenkins

## Usage

```go

import (
  "github.com/bndr/gojenkins"
  "context"
  "time"
  "fmt"
)

ctx := context.Background()
jenkins := gojenkins.CreateJenkins(nil, "http://localhost:8080/", "admin", "admin")
// Provide CA certificate if server is using self-signed certificate
// caCert, _ := ioutil.ReadFile("/tmp/ca.crt")
// jenkins.Requester.CACert = caCert
_, err := jenkins.Init(ctx)


if err != nil {
  panic("Something Went Wrong")
}

queueid, err := jenkins.BuildJob(ctx, "#jobname", nil)
if err != nil {
  panic(err)
}
build, err := jenkins.GetBuildFromQueueID(ctx, queueid)
if err != nil {
  panic(err)
}

// Wait for build to finish
for build.IsRunning(ctx) {
  time.Sleep(5000 * time.Millisecond)
  build.Poll(ctx)
}

fmt.Printf("build number %d with result: %v\n", build.GetBuildNumber(), build.GetResult())

```

API Reference: https://godoc.org/github.com/bndr/gojenkins

## Examples

For all of the examples below first create a jenkins object
```go
import "github.com/bndr/gojenkins"

jenkins, _ := gojenkins.CreateJenkins(nil, "http://localhost:8080/", "admin", "admin").Init(ctx)
```

or if you don't need authentication:

```go
jenkins, _ := gojenkins.CreateJenkins(nil, "http://localhost:8080/").Init(ctx)
```

you can also specify your own `http.Client` (for instance, providing your own SSL configurations):

```go
client := &http.Client{ ... }
jenkins, := gojenkins.CreateJenkins(client, "http://localhost:8080/").Init(ctx)
```

By default, `gojenkins` will use the `http.DefaultClient` if none is passed into the `CreateJenkins()`
function.

### Check Status of all nodes

```go
nodes := jenkins.GetAllNodes(ctx)

for _, node := range nodes {

  // Fetch Node Data
  node.Poll(ctx)
	if node.IsOnline(ctx) {
		fmt.Println("Node is Online")
	}
}

```

### Get all Builds for specific Job, and check their status

```go
jobName := "someJob"
builds, err := jenkins.GetAllBuildIds(ctx, jobName)

if err != nil {
  panic(err)
}

for _, build := range builds {
  buildId := build.Number
  data, err := jenkins.GetBuild(ctx, jobName, buildId)

  if err != nil {
    panic(err)
  }

	if "SUCCESS" == data.GetResult(ctx) {
		fmt.Println("This build succeeded")
	}
}

// Get Last Successful/Failed/Stable Build for a Job
job, err := jenkins.GetJob(ctx, "someJob")

if err != nil {
  panic(err)
}

job.GetLastSuccessfulBuild(ctx)
job.GetLastStableBuild(ctx)

```

### Get Current Tasks in Queue, and the reason why they're in the queue

```go

tasks := jenkins.GetQueue(ctx)

for _, task := range tasks {
	fmt.Println(task.GetWhy(ctx))
}

```

### Create View and add Jobs to it

```go

view, err := jenkins.CreateView(ctx, "test_view", gojenkins.LIST_VIEW)

if err != nil {
  panic(err)
}

status, err := view.AddJob(ctx, "jobName")

if status != nil {
  fmt.Println("Job has been added to view")
}

```

### Create nested Folders and create Jobs in them

```go

// Create parent folder
pFolder, err := jenkins.CreateFolder(ctx, "parentFolder")
if err != nil {
  panic(err)
}

// Create child folder in parent folder
cFolder, err := jenkins.CreateFolder(ctx, "childFolder", pFolder.GetName())
if err != nil {
  panic(err)
}

// Create job in child folder
configString := `<?xml version='1.0' encoding='UTF-8'?>
<project>
  <actions/>
  <description></description>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.scm.NullSCM"/>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers class="vector"/>
  <concurrentBuild>false</concurrentBuild>
  <builders/>
  <publishers/>
  <buildWrappers/>
</project>`

job, err := jenkins.CreateJobInFolder(ctx, configString, "jobInFolder", pFolder.GetName(), cFolder.GetName())
if err != nil {
  panic(err)
}

if job != nil {
	fmt.Println("Job has been created in child folder")
}

```

### Get All Artifacts for a Build and Save them to a folder

```go

job, _ := jenkins.GetJob(ctx, "job")
build, _ := job.GetBuild(ctx, 1)
artifacts := build.GetArtifacts(ctx)

for _, a := range artifacts {
	a.SaveToDir("/tmp")
}

```

### To always get fresh data use the .Poll() method

```go

job, _ := jenkins.GetJob(ctx, "job")
job.Poll()

build, _ := job.getBuild(ctx, 1)
build.Poll()

```

## Testing

    go test

## Contribute

All Contributions are welcome. The todo list is on the bottom of this README. Feel free to send a pull request.

## TODO

Although the basic features are implemented there are many optional features that are on the todo list.

* Kerberos Authentication
* CLI Tool
* Rewrite some (all?) iterators with channels

## LICENSE

Apache License 2.0
