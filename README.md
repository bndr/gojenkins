# Jenkins API Client for Go

[![GoDoc](https://godoc.org/github.com/bndr/gojenkins?status.svg)](https://godoc.org/github.com/bndr/gojenkins)
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

import "github.com/bndr/gojenkins"

jenkins := gojenkins.CreateJenkins("http://localhost:8080/", "admin", "admin").Init()

build := jenkins.GetJob("job_name").GetLastSuccessfulBuild()
duration := build.GetDuration()

job := jenkins.GetJob("jobname").Rename("SomeotherJobName")

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

j.CreateJob(configString, "someNewJobsName")


```

API Reference: https://godoc.org/github.com/bndr/gojenkins

## Examples

For all of the examples below first create a jenkins object
```go
import "github.com/bndr/gojenkins"

jenkins := gojenkins.CreateJenkins("http://localhost:8080/", "admin", "admin").Init()
```

or if you don't need authentication:

```go
jenkins := gojenkins.CreateJenkins("http://localhost:8080/").Init()
```

### Check Status of all nodes

```go
nodes := jenkins.GetAllNodes()

for _, node := range nodes {

  // Fetch Node Data
  node.Poll()
	if node.IsOnline() {
		fmt.Println("Node is Online")
	}
}

```

### Get all Builds for specific Job, and check their status

```go
jobName := "someJob"
builds := jenkins.GetAllBuildIds(jobName)

for _, build := range builds {
  buildId := build.Number
  data := jenkins.GetBuild(jobName, buildId)
	if "SUCCESS" == data.GetResult() {
		fmt.Println("This build succeeded")
	}
}

// Get Last Successful/Failed/Stable Build for a Job
jenkins.GetJob("someJob").GetLastSuccessfulBuild()
jenkins.GetJob("someJob").GetLastStableBuild()

```

### Get Current Tasks in Queue, and the reason why they're in the queue

```go

tasks := jenkins.GetQueue()

for _, task := range tasks {
	fmt.Println(task.GetWhy())
}

```

### Get All Artifacts for a Build and Save them to a folder

```go

job := jenkins.GetJob("job")
build := job.GetBuild(1)
artifacts := build.GetArtifacts()

for _, a := range artifacts {
	a.SaveToDir("/tmp")
}

```

### To always get fresh data use the .Poll() method

```go

job := jenkins.GetJob("job")
job.Poll()

build := job.getBuild(1)
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
