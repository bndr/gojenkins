package main

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
	Port      string
	Version   string
	BasicAuth *BasicAuth
	TokenAuth *TokenAuth
}

// Jenkins

func (j *Jenkins) connect() {

}

func (j *Jenkins) validate() {

}

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
