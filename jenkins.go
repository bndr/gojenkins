package main

type Jenkins struct {
}

type Job struct {
}

type Node struct {
}

type Queue struct {
}

type Build struct {
}

// Jenkins

func (j *Jenkins) Info() {

}
func (j *Jenkins) CreateNode() {

}
func (j *Jenkins) CreateBuild() {

}
func (j *Jenkins) CreateJob() {

}
func (j *Jenkins) GetNode() {

}
func (j *Jenkins) GetBuild() {

}
func (j *Jenkins) GetJob() {

}
func (j *Jenkins) GetAllNodes() {

}
func (j *Jenkins) GetAllBuilds() {

}
func (j *Jenkins) GetAllJobs() {

}

// Jobs
func (j *Job) Name() {

}

func (j *Job) Debug() {

}

func (j *Job) Build() {

}

func (j *Job) Enable() {

}

func (j *Job) Disable() {

}

func (j *Job) Delete() {

}

func (j *Job) Rename() {

}

func (j *Job) Exists() {

}
func (j *Job) Create() {

}

func (j *Job) Config() {

}

func (j *Job) BuildUrl() {

}

// Builds
func (b *Build) Info() {

}

func (b *Build) Stop() {

}

func (b *Build) Console() {

}

// Queue

func (q *Queue) Info() {

}

func (q *Queue) Cancel() {

}

func (q *Queue) Jobs() {

}

// Nodes

func (n *Node) Exists() {

}

func (n *Node) Delete() {

}

func (n *Node) Disable() {

}

func (n *Node) Enable() {

}

func (n *Node) Create() {

}
